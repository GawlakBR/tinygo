//go:build amken_max14

// This file contains the pin mappings for the Amken Max14 Intelligent Motion controller board.
//

package machine

const (
	// Onboard crystal oscillator frequency, in MHz.
	xoscFreq = 12 // MHz
)

// I2C Default pins on Raspberry Pico.
const (
	I2C0_SDA_PIN = GPIO32
	I2C0_SCL_PIN = GPIO33

	I2C1_SDA_PIN = GPIO38
	I2C1_SCL_PIN = GPIO39
)

// SPI default pins
const (
	// Default Serial Clock Bus 0 for SPI communications
	SPI0_SCK_PIN = GPIO42
	// Default Serial Out Bus 0 for SPI communications
	SPI0_SDO_PIN = GPIO43 // Tx
	// Default Serial In Bus 0 for SPI communications
	SPI0_SDI_PIN = GPIO40 // Rx

	// Default Serial Clock Bus 1 for SPI communications
	SPI1_SCK_PIN = GPIO2
	// Default Serial Out Bus 1 for SPI communications
	SPI1_SDO_PIN = GPIO3 // Tx
	// Default Serial In Bus 1 for SPI communications
	SPI1_SDI_PIN = GPIO0 // Rx
)

// UART pins
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

var DefaultUART = UART0

var StepperCS = [8]Pin{
	GPIO47, GPIO8, GPIO11, GPIO21,
	GPIO20, GPIO24, GPIO27, GPIO28,
}

const (
	COMM_CS_PIN = GPIO1
	CAN_CS_PIN  = GPIO4
	CAN_INT_PIN = GPIO5

	MOTOR1_CS = GPIO47
	MOTOR2_CS = GPIO8
	MOTOR3_CS = GPIO11
	MOTOR4_CS = GPIO21
	MOTOR5_CS = GPIO20
	MOTOR6_CS = GPIO24
	MOTOR7_CS = GPIO27
	MOTOR8_CS = GPIO28

	IO_EXP_RST_PIN  = GPIO31
	IO_EXP_INTA_PIN = GPIO7
	IO_EXP_INTB_PIN = GPIO6
	IO_EXP_SCL_PIN  = GPIO33
	IO_EXP_SDA_PIN  = GPIO32

	MOTOR1_DIR_PIN = GPIO9
	MOTOR2_DIR_PIN = GPIO13
	MOTOR3_DIR_PIN = GPIO15
	MOTOR4_DIR_PIN = GPIO17
	MOTOR5_DIR_PIN = GPIO19
	MOTOR6_DIR_PIN = GPIO23
	MOTOR7_DIR_PIN = GPIO25
	MOTOR8_DIR_PIN = GPIO30

	MOTOR1_STEP_PIN = GPIO10
	MOTOR2_STEP_PIN = GPIO12
	MOTOR3_STEP_PIN = GPIO14
	MOTOR4_STEP_PIN = GPIO16
	MOTOR5_STEP_PIN = GPIO18
	MOTOR6_STEP_PIN = GPIO22
	MOTOR7_STEP_PIN = GPIO26
	MOTOR8_STEP_PIN = GPIO29
)

const (
	A2D1 = ADC6
	A2D2 = ADC5
)

const (
	END_STOP1 = GPIO44
	END_STOP2 = GPIO41
	END_STOP3 = GPIO37
	END_STOP4 = GPIO36
	END_STOP5 = GPIO35
	END_STOP6 = GPIO34
)

// USB identifiers
const (
	usb_STRING_PRODUCT      = "Max14 Intelligent Motion Controller"
	usb_STRING_MANUFACTURER = "AmkenLLC"
)

var (
	usb_VID uint16 = 0x2E8A
	usb_PID uint16 = 0x7303
)
