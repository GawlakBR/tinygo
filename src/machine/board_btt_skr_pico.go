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
	X_DIR = GPIO10
	Y_DIR = GPIO5
	Z_DIR = ADC2
	E_DIR = GPIO13
)

// TMC stepper driver motor step
const (
	X_STEP = GPIO11
	Y_STEP = GPIO6
	Z_STEP = GPIO19
	E_STEP = GPIO14
)

// TMC stepper driver enable
const (
	X_ENABLE = GPIO12
	Y_ENABLE = GPIO7
	Z_ENABLE = GPIO2
	E_ENABLE = GPIO15
)

// TMC stepper driver UART
const (
	TMC_UART_TX = GPIO8
	TMC_UART_RX = GPIO9
)

// Endstops
const (
	X_ENDSTOP = GPIO4
	Y_ENDSTOP = GPIO3
	Z_ENDSTOP = GPIO25
	E_ENDSTOP = GPIO16
)

// Fan PWM
const (
	FAN1_PWM = GPIO17
	FAN2_PWM = GPIO18
	FAN3_PWM = GPIO20
)

// Heater PWM
const (
	HEATER_BED_PWM      = GPIO21
	HEATER_EXTRUDER_PWM = GPIO23
)

// Thermistors
const (
	THERM_BED      = ADC0 // Bed heater
	THERM_EXTRUDER = ADC1 // Toolhead heater
)

// Misc
const (
	RGB        = GPIO24 // Neopixel
	SERVO_ADC3 = ADC3   // Servo
	PROBE      = GPIO22 // Probe
)

// Onboard crystal oscillator frequency, in MHz.
const (
	xoscFreq = 12 // MHz
)

// I2C. We don't have this available, so map it to defaults
const (
	I2C0_SDA_PIN = GPIO4
	I2C0_SCL_PIN = GPIO5

	I2C1_SDA_PIN = GPIO2
	I2C1_SCL_PIN = GPIO3
)

// SPI default pins
const (
	SPI0_SCK_PIN = GPIO18
	SPI0_SDO_PIN = GPIO19 // Tx
	SPI0_SDI_PIN = GPIO16 // Rx
	SPI1_SCK_PIN = GPIO10
	SPI1_SDO_PIN = GPIO11 // Tx
	SPI1_SDI_PIN = GPIO12 // Rx
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
	UART1_TX_PIN = TMC_UART_TX
	UART1_RX_PIN = TMC_UART_RX
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

var DefaultUART = UART0
