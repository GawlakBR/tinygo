//go:build rp2040 || rp2350

package machine

import (
	"device/arm"
	"device/rp"
	"runtime/volatile"
	"unsafe"
)

var SysClockFrequency SysFrequency // in MHz
var xoscFreq xtalFreq

type SysFrequency uint32
type xtalFreq uint32

const (
	Freq8MHz   xtalFreq     = 8
	Freq12MHz  xtalFreq     = 12
	Freq20MHz  xtalFreq     = 20
	Freq24MHz  xtalFreq     = 24
	Freq25MHz  xtalFreq     = 25
	Freq40MHz  xtalFreq     = 40
	Freq48MHz  SysFrequency = 48  // USB
	Freq50MHz  SysFrequency = 50  // Good PWM Frequency
	Freq100MHz SysFrequency = 100 // Good PWM Frequency
	Freq125MHz SysFrequency = 125 // Valid frequency
	Freq133MHz SysFrequency = 133 // Valid frequency
	Freq150MHz SysFrequency = 150 // Valid frequency
	Freq175MHz SysFrequency = 175 // Valid frequency
	Freq200MHz SysFrequency = 200 // Good PWM Frequency
	Freq225MHz SysFrequency = 225 // Valid frequency
	Freq240MHz SysFrequency = 240 // Good PWM Frequency
	Freq250MHz SysFrequency = 250 // Good PWM Frequency
	Freq275MHz SysFrequency = 275 // Valid frequency
	Freq300MHz SysFrequency = 300 //  Good PWM Frequency
)

type PLLConfig struct {
	xoscFreq uint32
	sysFreq  SysFrequency
	refdiv   uint32
	fbdiv    uint32
	postDiv1 uint32
	postDiv2 uint32
}

// Flattened map with composite keys of frequencies drieved from SDK vcocalc.py
// [50,100,125,133,150,175,200,225,240,250,275,300].
// These frequencies are stable at core voltage of 1.1v
var pllConfigMap = map[[2]uint32]PLLConfig{
	{12, 48}:  {12, 48, 1, 120, 6, 5},
	{12, 50}:  {12, 50, 1, 125, 6, 5},
	{12, 100}: {12, 100, 1, 125, 5, 3},
	{12, 125}: {12, 125, 1, 125, 6, 2},
	{12, 133}: {12, 133, 1, 133, 6, 2},
	{12, 150}: {12, 150, 1, 125, 5, 2},
	{12, 175}: {12, 175, 2, 175, 6, 1},
	{12, 200}: {12, 200, 1, 100, 6, 1},
	{12, 225}: {12, 225, 2, 225, 6, 1},
	{12, 240}: {12, 240, 1, 120, 6, 1},
	{12, 250}: {12, 250, 1, 125, 6, 1},
	{12, 300}: {12, 300, 1, 125, 5, 1},
	{40, 48}:  {40, 48, 1, 36, 6, 5},
	{40, 50}:  {40, 50, 2, 75, 6, 5},
	{40, 100}: {40, 100, 1, 40, 4, 4},
	{40, 125}: {40, 125, 2, 75, 6, 2},
	{40, 133}: {40, 133, 4, 133, 5, 2},
	{40, 150}: {40, 150, 2, 75, 5, 2},
	{40, 200}: {40, 200, 1, 40, 4, 2},
	{40, 225}: {40, 225, 8, 315, 7, 1},
	{40, 240}: {40, 240, 1, 36, 6, 1},
	{40, 250}: {40, 250, 2, 75, 6, 1},
	{40, 300}: {40, 300, 2, 75, 5, 1},
}

func getClockConfig(xoscFreq, sysMHz uint32) (PLLConfig, uint32) {
	// Default configuration as fallback
	defaultConfig := PLLConfig{12, 48, 1, 120, 6, 5}
	cfg, exists := pllConfigMap[[2]uint32{xoscFreq, sysMHz}]
	if !exists {
		return defaultConfig, calculateVcoFreq(defaultConfig)
	}
	return cfg, calculateVcoFreq(cfg)
}

func calculateVcoFreq(cfg PLLConfig) uint32 {
	return (cfg.xoscFreq / cfg.refdiv) * cfg.fbdiv * MHz
}

func CPUFrequency() uint32 {
	return uint32(SysClockFrequency * MHz)
}

// Returns the period of a clock cycle for the raspberry pi pico in nanoseconds.
// Used in PWM API.
func cpuPeriod() uint32 {
	return 1e9 / CPUFrequency()
}

// clockIndex identifies a hardware clock
type clockIndex uint8

type clockType struct {
	ctrl     volatile.Register32
	div      volatile.Register32
	selected volatile.Register32
}

type fc struct {
	refKHz   volatile.Register32
	minKHz   volatile.Register32
	maxKHz   volatile.Register32
	delay    volatile.Register32
	interval volatile.Register32
	src      volatile.Register32
	status   volatile.Register32
	result   volatile.Register32
}

var clocks = (*clocksType)(unsafe.Pointer(rp.CLOCKS))

var configuredFreq [numClocks]uint32

type clock struct {
	*clockType
	cix clockIndex
}

// clock returns the clock identified by cix.
func (clks *clocksType) clock(cix clockIndex) clock {
	return clock{
		&clks.clk[cix],
		cix,
	}
}

// hasGlitchlessMux returns true if clock contains a glitchless multiplexer.
//
// Clock muxing consists of two components:
//
// A glitchless mux, which can be switched freely, but whose inputs must be
// free-running.
//
// An auxiliary (glitchy) mux, whose output glitches when switched, but has
// no constraints on its inputs.
//
// Not all clocks have both types of mux.
func (clk *clock) hasGlitchlessMux() bool {
	return clk.cix == clkSys || clk.cix == clkRef
}

// configure configures the clock by selecting the main clock source src
// and the auxiliary clock source auxsrc
// and finally setting the clock frequency to freq
// given the input clock source frequency srcFreq.
func (clk *clock) configure(src, auxsrc, srcFreq, freq uint32) {
	if freq > srcFreq {
		panic("clock frequency cannot be greater than source frequency")
	}

	div := calcClockDiv(srcFreq, freq)

	// If increasing divisor, set divisor before source. Otherwise set source
	// before divisor. This avoids a momentary overspeed when e.g. switching
	// to a faster source and increasing divisor to compensate.
	if div > clk.div.Get() {
		clk.div.Set(div)
	}

	// If switching a glitchless slice (ref or sys) to an aux source, switch
	// away from aux *first* to avoid passing glitches when changing aux mux.
	// Assume (!!!) glitchless source 0 is no faster than the aux source.
	if clk.hasGlitchlessMux() && src == rp.CLOCKS_CLK_SYS_CTRL_SRC_CLKSRC_CLK_SYS_AUX {
		clk.ctrl.ClearBits(rp.CLOCKS_CLK_REF_CTRL_SRC_Msk)
		for !clk.selected.HasBits(1) {
		}
	} else
	// If no glitchless mux, cleanly stop the clock to avoid glitches
	// propagating when changing aux mux. Note it would be a really bad idea
	// to do this on one of the glitchless clocks (clkSys, clkRef).
	{
		// Disable clock. On clkRef and ClkSys this does nothing,
		// all other clocks have the ENABLE bit in the same position.
		clk.ctrl.ClearBits(rp.CLOCKS_CLK_GPOUT0_CTRL_ENABLE_Msk)
		if configuredFreq[clk.cix] > 0 {
			// Delay for 3 cycles of the target clock, for ENABLE propagation.
			// Note XOSC_COUNT is not helpful here because XOSC is not
			// necessarily running, nor is timer... so, 3 cycles per loop:
			delayCyc := configuredFreq[clkSys]/configuredFreq[clk.cix] + 1
			for delayCyc != 0 {
				// This could be done more efficiently but TinyGo inline
				// assembly is not yet capable enough to express that. In the
				// meantime, this forces at least 3 cycles per loop.
				delayCyc--
				arm.Asm("nop\nnop\nnop")
			}
		}
	}

	// Set aux mux first, and then glitchless mux if this clock has one.
	clk.ctrl.ReplaceBits(auxsrc<<rp.CLOCKS_CLK_SYS_CTRL_AUXSRC_Pos,
		rp.CLOCKS_CLK_SYS_CTRL_AUXSRC_Msk, 0)

	if clk.hasGlitchlessMux() {
		clk.ctrl.ReplaceBits(src<<rp.CLOCKS_CLK_REF_CTRL_SRC_Pos,
			rp.CLOCKS_CLK_REF_CTRL_SRC_Msk, 0)
		for !clk.selected.HasBits(1 << src) {
		}
	}

	// Enable clock. On clkRef and clkSys this does nothing,
	// all other clocks have the ENABLE bit in the same position.
	clk.ctrl.SetBits(rp.CLOCKS_CLK_GPOUT0_CTRL_ENABLE)

	// Now that the source is configured, we can trust that the user-supplied
	// divisor is a safe value.
	clk.div.Set(div)

	// Store the configured frequency
	configuredFreq[clk.cix] = freq

}

// init initializes the clock hardware.
//
// Must be called before any other clock function.
func (clks *clocksType) init() {
	// Start the watchdog tick
	Watchdog.startTick(uint32(xoscFreq))

	// Disable resus that may be enabled from previous software
	rp.CLOCKS.SetCLK_SYS_RESUS_CTRL_CLEAR(0)

	// Enable the xosc
	xosc.init()

	// Before we touch PLLs, switch sys and ref cleanly away from their aux sources.
	clks.clk[clkSys].ctrl.ClearBits(rp.CLOCKS_CLK_SYS_CTRL_SRC_Msk)
	for !clks.clk[clkSys].selected.HasBits(0x1) {
	}

	clks.clk[clkRef].ctrl.ClearBits(rp.CLOCKS_CLK_REF_CTRL_SRC_Msk)
	for !clks.clk[clkRef].selected.HasBits(0x1) {
	}
	//// Configure PLLs
	cfg, vco := getClockConfig(uint32(xoscFreq), uint32(SysClockFrequency))
	pllSys.init(cfg.refdiv, vco, cfg.postDiv1, cfg.postDiv2)
	// Configure USB
	cfg, vco = getClockConfig(uint32(xoscFreq), 48)
	pllUSB.init(cfg.refdiv, vco, cfg.postDiv1, cfg.postDiv2)

	// We need to ensure clk_ref <= 25 MHz. We'll compute an integer divider:
	// minimal integer divisor so that (xoscFreq / refDiv) <= 25
	// i.e. refDiv >= xoscFreq / 25
	// We'll do a "ceiling" integer division: (xoscFreq + 24)/25
	refDiv := (xoscFreq + 24) / 25
	var refFreq = xoscFreq / refDiv
	// Configure clocks
	cref := clks.clock(clkRef)
	cref.configure(rp.CLOCKS_CLK_REF_CTRL_SRC_XOSC_CLKSRC,
		0, // No aux mux
		uint32(xoscFreq),
		uint32(refFreq))

	// Configure clkSys
	csys := clks.clock(clkSys)
	csys.configure(rp.CLOCKS_CLK_SYS_CTRL_SRC_CLKSRC_CLK_SYS_AUX,
		rp.CLOCKS_CLK_SYS_CTRL_AUXSRC_CLKSRC_PLL_SYS,
		uint32(SysClockFrequency*MHz),
		uint32(SysClockFrequency*MHz))

	// clkUSB = pllUSB (48MHz) / 1 = 48MHz
	cusb := clks.clock(clkUSB)
	cusb.configure(0, // No GLMUX
		rp.CLOCKS_CLK_USB_CTRL_AUXSRC_CLKSRC_PLL_USB,
		48*MHz,
		48*MHz)

	// clkADC = pllUSB (48MHZ) / 1 = 48MHz
	cadc := clks.clock(clkADC)
	cadc.configure(0, // No GLMUX
		rp.CLOCKS_CLK_ADC_CTRL_AUXSRC_CLKSRC_PLL_USB,
		48*MHz,
		48*MHz)

	clks.initRTC()

	// clkPeri = clkSys. Used as reference clock for Peripherals.
	// Cap at 150MHZ
	periFreq := SysClockFrequency
	if SysClockFrequency > 150 {
		periFreq = 150
	}

	cperi := clks.clock(clkPeri)
	cperi.configure(
		0, // no glitchless mux
		rp.CLOCKS_CLK_PERI_CTRL_AUXSRC_CLK_SYS,
		uint32(SysClockFrequency*MHz), // source is clk_sys
		uint32(periFreq*MHz),          // final freq is min(sysClkMHz, 150)
	)

	clks.initTicks()
}
