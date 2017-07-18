package main

const (
	ButtonA = iota
	ButtonB
	ButtonSelect
	ButtonStart
	ButtonUp
	ButtonDown
	ButtonLeft
	ButtonRight
)

type Controller struct {
	buttons [8]bool
	index int
	strobe byte
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) Read() byte {
	var data byte = 0
	if c.index < 8 && c.buttons[c.index] {
		data |= 1
	}
	if c.strobe & 1 == 1 {
		c.index = 0
	} else {
		c.index++
	}
	return data
	// XXX simulate open bus
}

func (c *Controller) Write(data byte) {
	c.strobe = data
	if c.strobe & 1 == 1 {
		c.index = 0
	}
}