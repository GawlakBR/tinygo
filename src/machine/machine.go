package machine

type GPIOConfig struct {
	Mode GPIOMode
}

type GPIO struct {
	Pin uint8
}

func (p GPIO) High() {
	p.Set(true)
}

func (p GPIO) Low() {
	p.Set(false)
}

type PWM struct {
	Pin uint8
}

type ADC struct {
	Pin uint8
}
