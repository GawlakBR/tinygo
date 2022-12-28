package main

import (
	"fmt"
	"machine"
)

var (
	audio = make([]int16, 16)
	pdm   = machine.PDM{}
)

func main() {
	machine.BUTTONA.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	err := pdm.Configure(machine.PDMConfig{})
	if err != nil {
		panic(fmt.Sprintf("Failed to configure PDM:%v", err))
	}

	for {
		if machine.BUTTONA.Get() {
			println("Recording new audio clip into memory")
			pdm.Read(&audio[0], uint32(len(audio)))
			println(fmt.Sprintf("Recorded new audio clip into memory: %v", audio))
		}
	}
}
