//go:build rp2040

// Implementation based on code located here:
// https://github.com/raspberrypi/pico-sdk/blob/master/src/rp2_common/hardware_rtc/rtc.c

package machine

import (
	"device/rp"
	"errors"
	"runtime/interrupt"
	"unsafe"
)

type rtcType rp.RTC_Type

var RTC = (*rtcType)(unsafe.Pointer(rp.RTC))

var rtcAlarmRepeats bool
var rtcCallback func()

var ErrRtcNoFreq = errors.New("RTC frequency not set")
var ErrRtcBigFreq = errors.New("RTC frequency too big to divide")
var ErrRtcNotActive = errors.New("RTC not active")
var ErrRtcInvalidTime = errors.New("invalid time for RTC")

func (rtc *rtcType) isActive() bool {
	return rtc.CTRL.HasBits(rp.RTC_CTRL_RTC_ACTIVE)
}

func (rtc *rtcType) init() error {
	// Get clk_rtc freq and make sure it is running
	rtcFreq := configuredFreq[clkRTC]
	if rtcFreq == 0 {
		return ErrRtcNoFreq
	}

	// Take rtc out of reset now that we know clk_rtc is running
	resetBlock(rp.RESETS_RESET_RTC)
	unresetBlockWait(rp.RESETS_RESET_RTC)

	// Set up the 1 second divider.
	// If rtc_freq is 400 then clkdiv_m1 should be 399
	rtcFreq -= 1

	// Check the freq is not too big to divide
	if rtcFreq > rp.RTC_CLKDIV_M1_CLKDIV_M1_Msk {
		return ErrRtcBigFreq
	}

	// Write divide value
	rtc.CLKDIV_M1.Set(rtcFreq)
	return nil
}

// SetTime configures RTC with supplied time, initialises and activates it.
func (rtc *rtcType) SetTime(t RtcTime) error {
	if !t.isValid() {
		return ErrRtcInvalidTime
	}

	err := rtc.init()
	if err != nil {
		return err
	}

	// Disable RTC and wait while it is still running
	rtc.CTRL.Set(0)
	for rtc.isActive() {
	}

	rtc.SETUP_0.Set((uint32(t.Year) << rp.RTC_SETUP_0_YEAR_Pos) |
		(uint32(t.Month) << rp.RTC_SETUP_0_MONTH_Pos) |
		(uint32(t.Day) << rp.RTC_SETUP_0_DAY_Pos))

	rtc.SETUP_1.Set((uint32(t.Dotw) << rp.RTC_SETUP_1_DOTW_Pos) |
		(uint32(t.Hour) << rp.RTC_SETUP_1_HOUR_Pos) |
		(uint32(t.Min) << rp.RTC_SETUP_1_MIN_Pos) |
		(uint32(t.Sec) << rp.RTC_SETUP_1_SEC_Pos))

	// Load setup values into RTC clock domain
	rtc.CTRL.SetBits(rp.RTC_CTRL_LOAD)

	// Enable RTC and wait for it to be running
	rtc.CTRL.SetBits(rp.RTC_CTRL_RTC_ENABLE)
	for !rtc.isActive() {
	}

	return nil
}

// GetTime returns current RTC time if RTC is active.
func (rtc *rtcType) GetTime() (t RtcTime, err error) {
	// Make sure RTC is running
	if !rtc.isActive() {
		return RtcTime{}, ErrRtcNotActive
	}

	// Note: RTC_0 should be read before RTC_1
	rtc_0 := rtc.RTC_0.Get()
	rtc_1 := rtc.RTC_1.Get()

	t = RtcTime{
		Dotw:  int8((rtc_0 & rp.RTC_RTC_0_DOTW_Msk) >> rp.RTC_RTC_0_DOTW_Pos),
		Hour:  int8((rtc_0 & rp.RTC_RTC_0_HOUR_Msk) >> rp.RTC_RTC_0_HOUR_Pos),
		Min:   int8((rtc_0 & rp.RTC_RTC_0_MIN_Msk) >> rp.RTC_RTC_0_MIN_Pos),
		Sec:   int8((rtc_0 & rp.RTC_RTC_0_SEC_Msk) >> rp.RTC_RTC_0_SEC_Pos),
		Year:  int16((rtc_1 & rp.RTC_RTC_1_YEAR_Msk) >> rp.RTC_RTC_1_YEAR_Pos),
		Month: int8((rtc_1 & rp.RTC_RTC_1_MONTH_Msk) >> rp.RTC_RTC_1_MONTH_Pos),
		Day:   int8((rtc_1 & rp.RTC_RTC_1_DAY_Msk) >> rp.RTC_RTC_1_DAY_Pos),
	}

	return t, nil
}

// SetAlarm configures alarm in RTC and arms it.
// The callback is executed in the context of an interrupt handler,
// so regular restructions for this sort of code apply: no blocking, no memory allocation, etc.
func (rtc *rtcType) SetAlarm(t RtcTime, callback func()) {

	rtc.disableInterruptMatch()

	// Clear all match enable bits
	rtc.IRQ_SETUP_0.ClearBits(rp.RTC_IRQ_SETUP_0_YEAR_ENA | rp.RTC_IRQ_SETUP_0_MONTH_ENA | rp.RTC_IRQ_SETUP_0_DAY_ENA)
	rtc.IRQ_SETUP_1.ClearBits(rp.RTC_IRQ_SETUP_1_DOTW_ENA | rp.RTC_IRQ_SETUP_1_HOUR_ENA | rp.RTC_IRQ_SETUP_1_MIN_ENA | rp.RTC_IRQ_SETUP_1_SEC_ENA)

	// Only add to setup if it isn't -1 and set the match enable bits for things we care about
	if t.Year >= 0 {
		rtc.IRQ_SETUP_0.SetBits(uint32(t.Year) << rp.RTC_SETUP_0_YEAR_Pos)
		rtc.IRQ_SETUP_0.SetBits(rp.RTC_IRQ_SETUP_0_YEAR_ENA)
	}

	if t.Month >= 0 {
		rtc.IRQ_SETUP_0.SetBits(uint32(t.Month) << rp.RTC_SETUP_0_MONTH_Pos)
		rtc.IRQ_SETUP_0.SetBits(rp.RTC_IRQ_SETUP_0_MONTH_ENA)
	}

	if t.Day >= 0 {
		rtc.IRQ_SETUP_0.SetBits(uint32(t.Day) << rp.RTC_SETUP_0_DAY_Pos)
		rtc.IRQ_SETUP_0.SetBits(rp.RTC_IRQ_SETUP_0_DAY_ENA)
	}

	if t.Dotw >= 0 {
		rtc.IRQ_SETUP_1.SetBits(uint32(t.Dotw) << rp.RTC_SETUP_1_DOTW_Pos)
		rtc.IRQ_SETUP_1.SetBits(rp.RTC_IRQ_SETUP_1_DOTW_ENA)
	}

	if t.Hour >= 0 {
		rtc.IRQ_SETUP_1.SetBits(uint32(t.Hour) << rp.RTC_SETUP_1_HOUR_Pos)
		rtc.IRQ_SETUP_1.SetBits(rp.RTC_IRQ_SETUP_1_HOUR_ENA)
	}

	if t.Min >= 0 {
		rtc.IRQ_SETUP_1.SetBits(uint32(t.Min) << rp.RTC_SETUP_1_MIN_Pos)
		rtc.IRQ_SETUP_1.SetBits(rp.RTC_IRQ_SETUP_1_MIN_ENA)
	}

	if t.Sec >= 0 {
		rtc.IRQ_SETUP_1.SetBits(uint32(t.Sec) << rp.RTC_SETUP_1_SEC_Pos)
		rtc.IRQ_SETUP_1.SetBits(rp.RTC_IRQ_SETUP_1_SEC_ENA)
	}

	rtcAlarmRepeats = t.alarmRepeats()

	// Store function pointer we can call later
	rtcCallback = callback

	// Enable the IRQ at the proc
	interrupt.New(rp.IRQ_RTC_IRQ, rtcHandleInterrupt).Enable()
	irqSet(rp.IRQ_RTC_IRQ, true)

	// Enable the IRQ at the peri
	rtc.INTE.Set(rp.RTC_INTE_RTC)

	rtc.enableInterruptMatch()
}

// ---

func (rtc *rtcType) enableInterruptMatch() {
	// Set matching and wait for it to be enabled
	rtc.IRQ_SETUP_0.SetBits(rp.RTC_IRQ_SETUP_0_MATCH_ENA)
	for !rtc.IRQ_SETUP_0.HasBits(rp.RTC_IRQ_SETUP_0_MATCH_ACTIVE) {
	}
}

func (rtc *rtcType) disableInterruptMatch() {
	// Disable matching and wait for it to stop being active
	rtc.IRQ_SETUP_0.ClearBits(rp.RTC_IRQ_SETUP_0_MATCH_ENA)
	for rtc.IRQ_SETUP_0.HasBits(rp.RTC_IRQ_SETUP_0_MATCH_ACTIVE) {
	}
}

// ---

func rtcHandleInterrupt(itr interrupt.Interrupt) {
	// Always disable the alarm to clear the current IRQ.
	// Even if it is a repeatable alarm, we don't want it to keep firing.
	// If it matches on a second it can keep firing for that second.
	RTC.disableInterruptMatch()

	if rtcAlarmRepeats {
		// If it is a repeatable alarm, re-enable the alarm.
		RTC.enableInterruptMatch()
	}

	// Call user callback function
	if rtcCallback != nil {
		rtcCallback()
	}
}

// ---

type RtcTime struct {
	Year  int16
	Month int8
	Day   int8
	Dotw  int8
	Hour  int8
	Min   int8
	Sec   int8
}

// isValid when fields are in ranges taken from RTC doc.
// Note when setting an RTC alarm these values are allowed to be -1 to say "don't match this value"
func (t RtcTime) isValid() bool {
	if !(t.Year >= 0 && t.Year <= 4095) {
		return false
	}
	if !(t.Month >= 1 && t.Month <= 12) {
		return false
	}
	if !(t.Day >= 1 && t.Day <= 31) {
		return false
	}
	if !(t.Dotw >= 0 && t.Dotw <= 6) {
		return false
	}
	if !(t.Hour >= 0 && t.Hour <= 23) {
		return false
	}
	if !(t.Min >= 0 && t.Min <= 59) {
		return false
	}
	if !(t.Sec >= 0 && t.Sec <= 59) {
		return false
	}
	return true
}

// alarmRepeats if any value is set to -1 since we don't match on that value in SetAlarm
func (t RtcTime) alarmRepeats() bool {
	return t.Year < 0 || t.Month < 0 || t.Day < 0 || t.Dotw < 0 || t.Hour < 0 || t.Min < 0 || t.Sec < 0
}
