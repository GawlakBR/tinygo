package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/keypad4x4"
)

func main() {
	keypad := keypad4x4.NewDevice(machine.D11, machine.D10, machine.D9, machine.D8, machine.D7, machine.D6, machine.D5, machine.D4)
	keypad.Configure()
	letters := [16]string{"1", "2", "3", "A", "4", "5", "6", "B", "7", "8", "9", "C", "*", "0", "#", "D"}

	ledPin := machine.D13
	ledPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	for {
		time.Sleep(50 * time.Millisecond) // debounce
		key := keypad.GetKey()
		if key == 255 { // value if nothing pressed
			continue
		}
		letter := letters[key]
		println("key", key, "letter", letter)

		switch letter {
		case "*":
			ledPin.Low()
		case "#":
			ledPin.High()
		default:
		}
	}
}
