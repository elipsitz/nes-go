package main

import "fmt"

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
	status_N bool // negative
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

func (cpu *Cpu6502) emulate(cycles int) {
	cycles_left := cycles
	for cycles_left > 0 {
		opcode := cpu.nes.read_byte(cpu.PC)
		fmt.Printf("reading from $%X : opcode $%X\n", cpu.PC, opcode)

		switch opcode {
		case 0x78:
			// SEI - Set Interrupt Disable
			cpu.status_I = true
			cpu.PC += 1
			cycles_left -= 2
		case 0xD8:
			// CLD - Clear Decimal Mode
			cpu.status_D = false
			cpu.PC += 1
			cycles_left -= 2
		case 0xA9:
			// LDA - Load Accumulator (Immediate)
			cpu.A = cpu.nes.read_byte(cpu.PC + 1)
			cpu.status_Z = cpu.A == 0
			cpu.status_N = (cpu.A & 0x80) > 0;
			cpu.PC += 2
			cycles_left -= 2
		case 0x8D:
			// STA - Store Accumulator (Absolute)

		default:
			panic(fmt.Sprintf("Unknown Opcode! $%X", opcode))
		}
	}
}