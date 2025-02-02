//go:build rp2350b

package machine

import (
	"errors"
)

const (
	adc0_CH ADCChannel = iota
	adc1_CH
	adc2_CH
	adc3_CH // Note: GPIO29 not broken out on pico board
	adc4_CH
	adc5_CH
	adc6_CH
	adc7_CH
	adcTempSensor // Internal temperature sensor channel
)

// GetADCChannel returns the channel associated with the ADC pin.
func (a ADC) GetADCChannel() (c ADCChannel, err error) {
	err = nil
	switch a.Pin {
	case ADC0:
		c = adc0_CH
	case ADC1:
		c = adc1_CH
	case ADC2:
		c = adc2_CH
	case ADC3:
		c = adc3_CH
	case ADC4:
		c = adc4_CH
	case ADC5:
		c = adc5_CH
	case ADC6:
		c = adc6_CH
	case ADC7:
		c = adc7_CH
	default:
		err = errors.New("no ADC channel for pin value")
	}
	return c, err
}

// The Pin method returns the GPIO Pin associated with the ADC mux channel, if it has one.
func (c ADCChannel) Pin() (p Pin, err error) {
	err = nil
	switch c {
	case adc0_CH:
		p = ADC0
	case adc1_CH:
		p = ADC1
	case adc2_CH:
		p = ADC2
	case adc3_CH:
		p = ADC3
	case adc4_CH:
		p = ADC4
	case adc5_CH:
		p = ADC5
	case adc6_CH:
		p = ADC6
	case adc7_CH:
		p = ADC7
	default:
		err = errors.New("no associated pin for channel")
	}
	return p, err
}
