//go:build rp2040 || rp2350a

package machine

const (
	maxPWMPins = 29
)

// Hardware Pulse Width Modulation (PWM) API
// PWM peripherals available on RP2040 and RP2350A. Each peripheral has 2 pins available for
// a total of 16 available PWM outputs. Some pins may not be available on some boards.
//
// The PWM block has 8 identical slices. Each slice can drive two PWM output signals, or
// measure the frequency or duty cycle of an input signal. This gives a total of up to 16 controllable
// PWM outputs. All 30 GPIOs can be driven by the PWM block
//
// The PWM hardware functions by continuously comparing the input value to a free-running counter. This produces a
// toggling output where the amount of time spent at the high output level is proportional to the input value. The fraction of
// time spent at the high signal level is known as the duty cycle of the signal.
//
// The default behaviour of a PWM slice is to count upward until the wrap value (\ref pwm_config_set_wrap) is reached, and then
// immediately wrap to 0. PWM slices also offer a phase-correct mode, where the counter starts to count downward after
// reaching TOP, until it reaches 0 again.
var (
	PWM0 = getPWMGroup(0)
	PWM1 = getPWMGroup(1)
	PWM2 = getPWMGroup(2)
	PWM3 = getPWMGroup(3)
	PWM4 = getPWMGroup(4)
	PWM5 = getPWMGroup(5)
	PWM6 = getPWMGroup(6)
	PWM7 = getPWMGroup(7)
)
