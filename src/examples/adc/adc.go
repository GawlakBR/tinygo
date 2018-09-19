package main

import (
	"machine"
	"time"
)

// This example assumes that an analog sensor such as a rotary dial is connected to pin ADC0.
// When the dial is turned past the midway point, the built-in LED will light up.

func main() {
	machine.InitADC()

	led := machine.GPIO{machine.LED}
	led.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})

	sensor := machine.ADC{machine.ADC2}
	sensor.Configure()

	for {
		val := sensor.Get()
		if val < 512 {
			led.Low()
		} else {
			led.High()
		}
		time.Sleep(time.Millisecond * 100)
	}
}
