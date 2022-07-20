package main

import (
	"fmt"
	"machine"
	"machine/usb/midi"
	"time"
)

func main() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	button := machine.BUTTON
	button.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	m := midi.New()
	m.SetHandler(func(b []byte) {
		led.Set(!led.Get())
		fmt.Printf("% X\r\n", b)
		m.Write(b)
	})

	prev := true
	chords := []struct {
		name string
		keys []byte
	}{
		{name: "C ", keys: []byte{60, 64, 67}},
		{name: "G ", keys: []byte{55, 59, 62}},
		{name: "Am", keys: []byte{57, 60, 64}},
		{name: "F ", keys: []byte{53, 57, 60}},
	}
	index := 0

	for {
		current := button.Get()
		if prev != current {
			led.Set(current)
			if current {
				for _, c := range chords[index].keys {
					m.NoteOff(0, 0, c, 0x40)
				}
				index = (index + 1) % len(chords)
			} else {
				for _, c := range chords[index].keys {
					m.NoteOn(0, 0, c, 0x40)
				}
			}
			prev = current
		}
		time.Sleep(10 * time.Millisecond)
	}
}
