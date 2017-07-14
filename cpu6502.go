package main

type address uint16

type Cpu6502 struct {
	nes *Nes

	A  byte    // Accumulator
	X  byte    // X index
	Y  byte    // Y index
	PC address // program counter
	SP byte    // stack pointer

	status_C bool // "Carry"
	status_Z bool // "Zero"
	status_I bool // Interrupt enable/disable
	status_D bool // BCD enable/disable
	status_B bool // BRK software interrupt
	status_V bool // Overflow
	status_S bool // sign
}

func NewCpu6502(nes *Nes) Cpu6502 {
	return Cpu6502{
		nes: nes,
		A: 0,
		X: 0,
		Y: 0,
		SP: 0xFD,
		status_I: true,
		status_B: true,
	}
}