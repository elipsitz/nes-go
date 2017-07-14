package main

import (
	"fmt"
	"time"
)

type address uint16

type Cpu struct {
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

func NewCpu(nes *Nes) Cpu {
	return Cpu{
		nes: nes,
		A: 0,
		X: 0,
		Y: 0,
		SP: 0xFD,
		status_I: true,
		status_B: true,
	}
}

func (cpu *Cpu) Emulate(cycles int) {
	cycles_left := cycles
	for cycles_left > 0 {
		opcode := cpu.nes.read_byte(cpu.PC)
		// fmt.Printf("reading from $%X : opcode $%X\n", cpu.PC, opcode)

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
		case 0xAD:
			// LDA - Load Accumulator (Absolute)
			addr := address(cpu.nes.read_uint16(cpu.PC + 1))
			cpu.A = cpu.nes.read_byte(addr)
			cpu.status_Z = cpu.A == 0
			cpu.status_N = (cpu.A & 0x80) > 0;
			cpu.PC += 3
			cycles_left -= 4
		case 0x8D:
			// STA - Store Accumulator (Absolute)
			addr := address(cpu.nes.read_uint16(cpu.PC + 1))
			cpu.nes.write_byte(addr, cpu.A)
			cpu.PC += 3
			cycles_left -= 4
		case 0xA2:
			// LDX - Load X Register (Immediate)
			cpu.X = cpu.nes.read_byte(cpu.PC + 1)
			cpu.status_Z = cpu.X == 0
			cpu.status_N = (cpu.X & 0x80) > 0;
			cpu.PC += 2
			cycles_left -= 2
		case 0x9A:
			// TXS - Transfer X to Stack Pointer
			cpu.SP = cpu.X
			cpu.PC += 1
			cycles_left -= 2
		case 0x29:
			// AND - Logical AND (Immediate)
			data := cpu.nes.read_byte(cpu.PC + 1)
			cpu.A = cpu.A & data
			cpu.status_Z = cpu.A == 0
			cpu.status_N = (cpu.A & 0x80) > 0;
			cpu.PC += 2
			cycles_left -= 2
		case 0xF0:
			// BEQ - Branch if Equal
			if cpu.status_Z {
				var offset int8 = int8(cpu.nes.read_byte(cpu.PC + 1))
				cpu.PC = address(int32(cpu.PC) + int32(offset))
			}
			cpu.PC += 2
			cycles_left -= 2 // XXX CHECK THIS
		default:
			time.Sleep(100000000)
			panic(fmt.Sprintf("Unknown Opcode! $%X", opcode))
		}
	}
}