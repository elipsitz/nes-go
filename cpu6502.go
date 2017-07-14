package main

type Cpu6502 struct {
	A uint8 // Accumulator
	X uint8 // X index
	Y uint8 // Y index

	status_C bool // "Carry"
	status_Z bool // "Zero"
	status_I bool // Interrupt enable/disable
	status_D bool // BCD enable/disable
	status_B bool // BRK software interrupt
	status_V bool // Overflow
	status_S bool // sign

	PC uint16 // program counter
	SP uint8  // stack pointer
}
