//go:build rp2350

package machine

import (
	"device/rp"
	"runtime/volatile"
	"unsafe"
)

// On some boards, the XOSC can take longer than usual to stabilize. On such
// boards, this is needed to avoid a hard fault on boot/reset. Refer to
// PICO_XOSC_STARTUP_DELAY_MULTIPLIER in the Pico SDK for additional details.
const XOSC_STARTUP_DELAY_MULTIPLIER = 64

type xoscType struct {
	ctrl     volatile.Register32
	status   volatile.Register32
	dormant  volatile.Register32
	startup  volatile.Register32
	reserved [3 - 3*rp2350ExtraReg]volatile.Register32
	count    volatile.Register32
}

var xosc = (*xoscType)(unsafe.Pointer(rp.XOSC))

// init initializes the crystal oscillator system.
//
// This function will block until the crystal oscillator has stabilised.
func (osc *xoscType) init() {

	// Choose the correct FREQ_RANGE value from the enumerations.
	// Note: these ranges come from the RP2 datasheet’s "XOSC_CTRL FREQ_RANGE" table:
	//   0xaa0 => 1..15 MHz
	//   0xaa1 => 10..30 MHz
	//   0xaa2 => 25..60 MHz
	//   0xaa3 => 40..100 MHz
	var ctrlVal uint32
	if xoscFreq <= 15 {
		ctrlVal = rp.XOSC_CTRL_FREQ_RANGE_1_15MHZ
	} else if xoscFreq <= 30 {
		ctrlVal = rp.XOSC_CTRL_FREQ_RANGE_10_30MHZ
	} else if xoscFreq <= 60 {
		ctrlVal = rp.XOSC_CTRL_FREQ_RANGE_25_60MHZ
	} else if xoscFreq <= 100 {
		ctrlVal = rp.XOSC_CTRL_FREQ_RANGE_40_100MHZ
	} else {
		panic("unsupported freq")
	}
	osc.ctrl.Set(ctrlVal)
	// Set xosc startup delay
	delay := (((xoscFreq * MHz) / 1000) + 128) / 256 * XOSC_STARTUP_DELAY_MULTIPLIER
	osc.startup.Set(uint32(delay))

	// Set the enable bit now that we have set freq range and startup delay
	osc.ctrl.SetBits(rp.XOSC_CTRL_ENABLE_ENABLE << rp.XOSC_CTRL_ENABLE_Pos)

	// Wait for xosc to be stable
	for !osc.status.HasBits(rp.XOSC_STATUS_STABLE) {
	}
}
