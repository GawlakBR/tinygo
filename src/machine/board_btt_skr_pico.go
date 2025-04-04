//go:build btt_skr_pico

// This contains the pin mappings for the BigTreeTech SKR Pico.
//
// Purchase link: https://biqu.equipment/products/btt-skr-pico-v1-0
// Board schematic: https://github.com/bigtreetech/SKR-Pico/blob/master/Hardware/BTT%20SKR%20Pico%20V1.0-SCH.pdf
// Pin diagram: https://github.com/bigtreetech/SKR-Pico/blob/master/Hardware/BTT%20SKR%20Pico%20V1.0-PIN.pdf

package machine

// TMC stepper driver motor direction.
// X/Y/Z/E refers to motors for X/Y/Z and the extruder.
const (
	X_DIR Pin = GPIO10
	Y_DIR Pin = GPIO5
	Z_DIR Pin = ADC2
	E_DIR Pin = GPIO13
)

// TMC stepper driver motor step
const (
	X_STEP Pin = GPIO11
	Y_STEP Pin = GPIO6
	Z_STEP Pin = GPIO19
	E_STEP Pin = GPIO14
)

// TMC stepper driver enable
const (
	X_ENABLE Pin = GPIO12
	Y_ENABLE Pin = GPIO7
	Z_ENABLE Pin = GPIO2
	E_ENABLE Pin = GPIO15
)

// TMC stepper driver UART
const (
	TMC_UART_TX Pin = GPIO8
	TMC_UART_RX Pin = GPIO9
)

// Endstops
const (
	X_ENDSTOP Pin = GPIO4
	Y_ENDSTOP Pin = GPIO3
	Z_ENDSTOP Pin = GPIO25
	E_ENDSTOP Pin = GPIO16
)

// Fan PWM
const (
	FAN1_PWM Pin = GPIO17
	FAN2_PWM Pin = GPIO18
	FAN3_PWM Pin = GPIO20
)

// Heater PWM
const (
	HEATER_BED_PWM      Pin = GPIO21
	HEATER_EXTRUDER_PWM Pin = GPIO23
)

// Thermistors
const (
	THERM_BED          = ADC0 // Bed heater
	THERM_EXTRUDER Pin = ADC1 // Toolhead heater
)

// Misc
const (
	RGB        Pin = GPIO24 // Neopixel
	SERVO_ADC3 Pin = ADC3   // Servo
	PROBE      Pin = GPIO22 // Probe
)

// Onboard crystal oscillator frequency, in MHz.
const (
	xoscFreq = 12 // MHz
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "SKR Pico"
	usb_STRING_MANUFACTURER = "BigTreeTech"
)

var (
	usb_VID uint16 = 0x2e8a
	usb_PID uint16 = 0x0003
)

// UART pins
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

var DefaultUART = UART0
