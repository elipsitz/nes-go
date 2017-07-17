package main

import (
	"fmt"
	"time"
)

type Cpu struct {
	nes *Nes
	mem Memory

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

	pendingNmiInterrupt bool
}

func NewCpu(nes *Nes) *Cpu {
	return &Cpu{
		nes: nes,
		mem: &CPUMemory{nes: nes},
		A: 0,
		X: 0,
		Y: 0,
		SP: 0xFD,
		status_I: true,
		status_B: true,
	}
}

// interrupt vectors
func (cpu *Cpu) getVectorReset() address {
	return address(ReadUint16(cpu.mem, 0xFFFC))
}

func (cpu *Cpu) getVectorNMI() address {
	return address(ReadUint16(cpu.mem, 0xFFFA))
}

func (cpu *Cpu) getVectorBRK() address {
	return address(ReadUint16(cpu.mem, 0xFFFE))
}

// these return (address, instruction size in bytes, [page crossed?])
func (cpu *Cpu) addressImmediate() (address, int) {
	return cpu.PC + 1, 2
}

func (cpu *Cpu) addressZeroPage() (address, int) {
	return address(cpu.mem.Read(cpu.PC + 1)), 2
}

func (cpu *Cpu) addressZeroPageX() (address, int) {
	return address((cpu.mem.Read(cpu.PC + 1) + cpu.X) & 0xFF), 2
}

func (cpu *Cpu) addressZeroPageY() (address, int) {
	return address((cpu.mem.Read(cpu.PC + 1) + cpu.Y) & 0xFF), 2
}

func (cpu *Cpu) addressRelative() (address, int) {
	var offset int8 = int8(cpu.mem.Read(cpu.PC + 1))
	return address(int32(cpu.PC) + int32(offset)), 2
}

func (cpu *Cpu) addressAbsolute() (address, int) {
	return address(ReadUint16(cpu.mem, cpu.PC + 1)), 3
}

func (cpu *Cpu) addressAbsoluteX() (address, int, bool) {
	addr := address(ReadUint16(cpu.mem, cpu.PC + 1) + uint16(cpu.X))

	pageCrossed := false
	if uint16(addr & 0xFF) + uint16(cpu.X) > 255 {
		cpu.mem.Read(addr)
		pageCrossed = true
	}
	return addr, 3, pageCrossed
}

func (cpu *Cpu) addressAbsoluteY() (address, int, bool) {
	addr := address(ReadUint16(cpu.mem, cpu.PC + 1) + uint16(cpu.Y))

	pageCrossed := false
	if uint16(addr & 0xFF) + uint16(cpu.X) > 255 {
		cpu.mem.Read(addr)
		pageCrossed = true
	}
	return addr, 3, pageCrossed
}

func (cpu *Cpu) addressIndirectX() (address, int) {
	addr := address((cpu.mem.Read(cpu.PC + 1) + cpu.X) & 0xFF)
	return address(cpu.read_uint16_buggy(addr)), 2
}

func (cpu *Cpu) addressIndirectY() (address, int, bool) {
	addr := address(cpu.mem.Read(cpu.PC + 1))
	addr = address(cpu.read_uint16_buggy(addr) + uint16(cpu.Y))

	pageCrossed := false
	if uint16(addr & 0xFF) + uint16(cpu.Y) > 255 {
		cpu.mem.Read(addr) // ? is this right?
		pageCrossed = true
	}
	return addr, 2, pageCrossed
}

func (cpu *Cpu) stackPush(data byte) {
	cpu.mem.Write(address(0x0100 + uint16(cpu.SP)), data)
	cpu.SP--
}

func (cpu *Cpu) stackPull() (byte) {
	cpu.SP++
	return cpu.mem.Read(address(0x0100 + uint16(cpu.SP)))
}

func (cpu *Cpu) statusPack() (data byte) {
	if cpu.status_C {
		data |= 1 << 0
	}
	if cpu.status_Z {
		data |= 1 << 1
	}
	if cpu.status_I {
		data |= 1 << 2
	}
	if cpu.status_D {
		data |= 1 << 3
	}
	if cpu.status_B {
		// XXX note discussion of 'B' flag here: https://wiki.nesdev.com/w/index.php/Status_flags
		data |= 1 << 4
	}
	data |= 1 << 5
	if cpu.status_V {
		data |= 1 << 6
	}
	if cpu.status_N {
		data |= 1 << 7
	}
	return
}

func (cpu *Cpu) statusUnpack(data byte) {
	cpu.status_C = data & (1 << 0) > 0
	cpu.status_Z = data & (1 << 1) > 0
	cpu.status_I = data & (1 << 2) > 0
	cpu.status_D = data & (1 << 3) > 0
	cpu.status_B = data & (1 << 4) > 0 // XXX see above
	cpu.status_V = data & (1 << 6) > 0
	cpu.status_N = data & (1 << 7) > 0
}

func (cpu *Cpu) read_uint16_buggy(addr address) uint16 {
	low := cpu.mem.Read(addr)
	high := cpu.mem.Read((addr & 0xFF00) | address(byte(addr) + 1))
	return uint16(high) << 8 | uint16(low)
}

func (cpu *Cpu) handleInterrupt(addr address) {
	cpu.stackPush(byte((cpu.PC >> 8) & 0xFF))
	cpu.stackPush(byte(cpu.PC & 0xFF))
	cpu.stackPush(cpu.statusPack())
	cpu.PC = addr
}

// emulate for at least `cycles` cycles -- returns number of cycles actually emulated for
func (cpu *Cpu) Emulate(cycles int) int {
	cycles_left := cycles
	for cycles_left > 0 {
		if cpu.pendingNmiInterrupt {
			cpu.pendingNmiInterrupt = false
			cpu.handleInterrupt(cpu.getVectorNMI())
		}
		opcode := cpu.mem.Read(cpu.PC)

		// logline(fmt.Sprintf("%.4X  %.2X________________________________________A:%.2X X:%.2X Y:%.2X P:%.2X SP:%.2X CYC:___", cpu.PC, opcode, cpu.A, cpu.X, cpu.Y, cpu.statusPack() & 0xEF, cpu.SP))

		if opcode & 0x3 == 1 {
			var size, cycles int
			var addr address
			addressType, instructionType := (opcode >> 2) & 0x7, (opcode >> 5) & 0x7
			switch addressType {
			case 0:
				addr, size = cpu.addressIndirectX()
				cycles = 6
			case 1:
				addr, size = cpu.addressZeroPage()
				cycles = 3
			case 2:
				addr, size = cpu.addressImmediate()
				cycles = 2
			case 3:
				addr, size = cpu.addressAbsolute()
				cycles = 4
			case 4:
				var pageCrossed bool
				addr, size, pageCrossed = cpu.addressIndirectY()
				cycles = 5
				if pageCrossed || instructionType == 4 {
					cycles++
				}
			case 5:
				addr, size = cpu.addressZeroPageX()
			case 6:
				var pageCrossed bool
				addr, size, pageCrossed = cpu.addressAbsoluteY()
				cycles = 4
				if pageCrossed || instructionType == 4 {
					cycles++
				}
			case 7:
				var pageCrossed bool
				addr, size, pageCrossed = cpu.addressAbsoluteX()
				cycles = 4
				if pageCrossed || instructionType == 4 {
					cycles++
				}
			}

			switch instructionType {
			case 0:
				// ORA - Logical Inclusive OR
				cpu.A |= cpu.mem.Read(addr)
			case 1:
				// AND - Logical AND
				cpu.A &= cpu.mem.Read(addr)
			case 2:
				// EOR - Exclusive OR
				cpu.A ^= cpu.mem.Read(addr)
			case 3:
				// ADC - Add with Carry
				var op1, op2 byte = cpu.A, cpu.mem.Read(addr)
				var val uint16 = uint16(op1) + uint16(op2)
				if cpu.status_C {
					val += 1
				}
				cpu.status_C = val > 255
				cpu.A = byte(val)
				cpu.status_V = ((op1 & 0x80 > 0) && (op2 & 0x80 > 0) && (cpu.A & 0x80 == 0)) ||  ((op1 & 0x80 == 0) && (op2 & 0x80 == 0) && (cpu.A & 0x80 > 0))
			case 4:
				// STA - Store Accumulator
				cpu.mem.Write(addr, cpu.A)
			case 5:
				// LDA - Load Accumulator
				cpu.A = cpu.mem.Read(addr)
			case 6:
				// CMP - Compare
				data := cpu.mem.Read(addr)
				cpu.status_C = cpu.A >= data
				cpu.status_Z = cpu.A == data
				cpu.status_N = (cpu.A - data) & 0x80 > 0
			case 7:
				// SBC - Subtract With Carry
				var sub int8 = int8(cpu.mem.Read(addr))
				if !cpu.status_C {
					sub++
				}
				var op1, op2 byte = cpu.A, byte(-sub)
				var val uint16 = uint16(op1) + uint16(op2)

				cpu.status_C = val > 255
				cpu.A = byte(val)
				cpu.status_V = ((op1 & 0x80 > 0) && (op2 & 0x80 > 0) && (cpu.A & 0x80 == 0)) ||  ((op1 & 0x80 == 0) && (op2 & 0x80 == 0) && (cpu.A & 0x80 > 0))
			}

			if instructionType != 4 && instructionType != 6 {
				cpu.status_Z = cpu.A == 0
				cpu.status_N = (cpu.A & 0x80) > 0
			}

			cpu.PC += address(size)
			cycles_left -= cycles
			continue
		}

		// misc instructions
		switch opcode {
		case 0x00:
			// BRK - Force Interrupt
			returnAddr := cpu.PC + 1
			cpu.stackPush(byte((returnAddr >> 8) & 0xFF))
			cpu.stackPush(byte(returnAddr & 0xFF))
			cpu.stackPush(cpu.statusPack())
			cpu.status_B = true
			cycles_left -= 7
			cpu.PC = cpu.getVectorBRK()
		case 0x20:
			// JSR - Jump to Subroutine
			addr := address(ReadUint16(cpu.mem, cpu.PC + 1))
			returnAddr := cpu.PC + 3 - 1 // returnAddr minus one
			cpu.stackPush(byte((returnAddr >> 8) & 0xFF))
			cpu.stackPush(byte(returnAddr & 0xFF))
			cycles_left -= 6
			cpu.PC = addr
		case 0x40:
			// RTI - Return from Interrupt
			cpu.statusUnpack(cpu.stackPull())
			cpu.PC = address(cpu.stackPull()) + address(cpu.stackPull()) * 256
			cycles_left -= 6
		case 0x60:
			// RTS - Return from Subroutine
			cpu.PC = (address(cpu.stackPull()) + address(cpu.stackPull()) * 256) + 1
			cycles_left -= 6
		case 0x08:
			// PHP - Push Processor Status
			cpu.stackPush(cpu.statusPack())
			cpu.PC += 1
			cycles_left -= 3
		case 0x28:
			// PLP - Pull Processor Status
			cpu.statusUnpack(cpu.stackPull())
			cpu.PC += 1
			cycles_left -= 4
		case 0x48:
			// PHA - Push Accumulator
			cpu.stackPush(cpu.A)
			cpu.PC += 1
			cycles_left -= 3
		case 0x68:
			// PLA - Pull Accumulator
			cpu.A = cpu.stackPull()
			cpu.status_Z = cpu.A == 0
			cpu.status_N = (cpu.A & 0x80) > 0
			cpu.PC += 1
			cycles_left -= 4
		case 0x88:
			// DEY - Decrement Y Register
			cpu.Y -= 1
			cpu.status_Z = cpu.Y == 0
			cpu.status_N = (cpu.Y & 0x80) > 0
			cpu.PC += 1
			cycles_left -= 2
		case 0xA8:
			// TAY - Transfer Accumulator to Y
			cpu.Y = cpu.A
			cpu.status_Z = cpu.Y == 0
			cpu.status_N = (cpu.Y & 0x80) > 0
			cpu.PC += 1
			cycles_left -= 2
		case 0xC8:
			// INY - Increment Y Register
			cpu.Y += 1
			cpu.status_Z = cpu.Y == 0
			cpu.status_N = (cpu.Y & 0x80) > 0
			cpu.PC += 1
			cycles_left -= 2
		case 0xE8:
			// INX - Increment X Register
			cpu.X += 1
			cpu.status_Z = cpu.X == 0
			cpu.status_N = (cpu.X & 0x80) > 0
			cpu.PC += 1
			cycles_left -= 2
		case 0x18:
			// CLC - Clear Carry Flag
			cpu.status_C = false
			cpu.PC += 1
			cycles_left -= 2
		case 0x38:
			// CLC - Set Carry Flag
			cpu.status_C = true
			cpu.PC += 1
			cycles_left -= 2
		case 0x58:
			// CLI - Clear Interrupt Disable
			cpu.status_I = false
			cpu.PC += 1
			cycles_left -= 2
		case 0x78:
			// SEI - Set Interrupt Disable
			cpu.status_I = true
			cpu.PC += 1
			cycles_left -= 2
		case 0x98:
			// TYA - Transfer Y to Accumulator
			cpu.A = cpu.Y
			cpu.status_Z = cpu.A == 0
			cpu.status_N = (cpu.A & 0x80) > 0
			cpu.PC += 1
			cycles_left -= 2
		case 0xB8:
			// CLV - Clear Overflow Flag
			cpu.status_V = false
			cpu.PC += 1
			cycles_left -= 2
		case 0xD8:
			// CLD - Clear Decimal Mode
			cpu.status_D = false
			cpu.PC += 1
			cycles_left -= 2
		case 0xF8:
			// SED - Set Decimal Flag
			cpu.status_D = true
			cpu.PC += 1
			cycles_left -= 2
		case 0x9A:
			// TXS - Transfer X to Stack Pointer
			cpu.SP = cpu.X
			cpu.PC += 1
			cycles_left -= 2
		case 0xBA:
			// TSX - Transfer Stack Pointer to X
			cpu.X = cpu.SP
			cpu.status_Z = cpu.X == 0
			cpu.status_N = (cpu.X & 0x80) > 0
			cpu.PC += 1
			cycles_left -= 2
		case 0xCA:
			// DEY - Decrement Y Register
			cpu.X -= 1
			cpu.status_Z = cpu.X == 0
			cpu.status_N = (cpu.X & 0x80) > 0
			cpu.PC += 1
			cycles_left -= 2
		case 0xEA:
			// NOP - No Operation
			cpu.PC += 1
			cycles_left -= 2

		// HERE LIES UNDOCUMENTED OPCODES
		case 0x04:
			fallthrough
		case 0x44:
			fallthrough
		case 0x64:
			// 2 byte, 3 cycle NOP
			cpu.PC += 2
			cycles_left -= 3
		case 0x0C:
			fallthrough
		case 0x1C:
			fallthrough
		case 0x3C:
			fallthrough
		case 0x5C:
			fallthrough
		case 0x7C:
			fallthrough
		case 0xDC:
			fallthrough
		case 0xFC:
			// 3 bytes, 4 cycle NOP
			cpu.PC += 3
			cycles_left -= 4
		case 0x14:
			fallthrough
		case 0x34:
			fallthrough
		case 0x54:
			fallthrough
		case 0x74:
			fallthrough
		case 0xD4:
			fallthrough
		case 0xF4:
			// 2 byte, 4 cycle NOP
			cpu.PC += 2
			cycles_left -= 4
		case 0x1A:
			fallthrough
		case 0x3A:
			fallthrough
		case 0x5A:
			fallthrough
		case 0x7A:
			fallthrough
		case 0xDA:
			fallthrough
		case 0xFA:
			// 1 byte, 2 cycle NOP
			cpu.PC += 1
			cycles_left -= 2

		default:
			goto keep_going
		}

		continue
		keep_going:

		if opcode & 0x3 == 2 {
			var size, cycles int
			var addr address
			addressType, instructionType := (opcode >> 2) & 0x7, (opcode >> 5) & 0x7
			if addressType != 4 && addressType != 6 {
				switch addressType {
				case 0:
					addr, size = cpu.addressImmediate()
					cycles = 2
				case 1:
					addr, size = cpu.addressZeroPage()
					if instructionType != 4 && instructionType != 5 {
						cycles = 5
					} else {
						cycles = 3
					}
				case 2:
					// ACCUMULATOR!!!
					cycles = 2
					size = 1
				case 3:
					addr, size = cpu.addressAbsolute()
					if instructionType != 4 && instructionType != 5 {
						cycles = 6
					} else {
						cycles = 4
					}
				case 5:
					if instructionType != 4 && instructionType != 5 {
						addr, size = cpu.addressZeroPageX()
						cycles = 6
					} else {
						addr, size = cpu.addressZeroPageY()
						cycles = 4
					}
				case 7:
					var pageCrossed bool
					if instructionType != 4 && instructionType != 5 {
						addr, size, pageCrossed = cpu.addressAbsoluteX()
						cycles = 7
					} else {
						addr, size, pageCrossed = cpu.addressAbsoluteY()
						cycles = 4
						if pageCrossed {
							cycles++
						}
					}
				}

				var data byte
				if addressType != 2 {
					data = cpu.mem.Read(addr)
				} else {
					data = cpu.A
				}

				switch instructionType {
				case 0:
					// ASL - Arithmetic Shift Left
					cpu.status_C = (data & 0x80) > 0
					data = data << 1
				case 1:
					// ROL - Rotate Left
					oldCarry := cpu.status_C
					cpu.status_C = (data & 0x80) > 0
					data = data << 1
					if oldCarry {
						data |= 1
					}
				case 2:
					// LSR - Logical Shift Right
					cpu.status_C = (data & 0x1) > 0
					data = data >> 1
				case 3:
					// ROR - Rotate Right
					oldCarry := cpu.status_C
					cpu.status_C = (data & 0x1) > 0
					data = data >> 1
					if oldCarry {
						data |= 0x80
					}
				case 4:
					// STX - Store X Register
					// (also TXA when mode = accumulator)
					// no N/Z
					data = cpu.X
					if addressType == 2 {
						size = 1
					}
				case 5:
					// LDX - Load X Register
					// (also TAX when mode = accumulator)
					cpu.X = data
					if addressType == 2 {
						size = 1
					}
				case 6:
					// DEC - Decrement Memory
					data--
				case 7:
					// INC - Increment Memory
					data++
				}

				if addressType != 2 {
					if instructionType != 5 {
						cpu.mem.Write(addr, data)
					}
				} else {
					cpu.A = data
				}

				if instructionType != 4 || addressType == 2 {
					cpu.status_Z = data == 0
					cpu.status_N = (data & 0x80) > 0
				}

				cpu.PC += address(size)
				cycles_left -= cycles
				continue
			}
		}

		if opcode & 0x03 == 0 {
			var size, cycles int
			var addr address
			addressType, instructionType := (opcode >> 2) & 0x7, (opcode >> 5) & 0x7
			if instructionType != 0 && addressType != 2 && addressType != 4 && addressType != 6 {
				switch addressType {
				case 0:
					addr, size = cpu.addressImmediate()
					cycles = 2
				case 1:
					addr, size = cpu.addressZeroPage()
					cycles = 3
				case 3:
					if instructionType != 3 {
						addr, size = cpu.addressAbsolute()
						cycles = 4
						if instructionType == 2 {
							cycles = 3
						}
					} else {
						// jump indirect
						addr, size = address(ReadUint16(cpu.mem, cpu.PC + 1)), 3
						addr = address(cpu.read_uint16_buggy(addr))
						cycles = 5
					}
				case 5:
					addr, size = cpu.addressZeroPageX()
					cycles = 4
				case 7:
					var pageCrossed bool
					addr, size, pageCrossed = cpu.addressAbsoluteX()
					cycles = 4
					if pageCrossed {
						cycles += 1
					}
				}

				switch instructionType {
				case 1:
					// BIT - Bit Test
					data := cpu.mem.Read(addr)
					cpu.status_Z = (cpu.A & data) == 0
					cpu.status_V = (data & 0x40) > 0
					cpu.status_N = (data & 0x80) > 0
				case 2:
					// JMP() [absolute] - Jump
					cpu.PC = addr
				case 3:
					// JMP [indirect] - Jump
					cpu.PC = addr
				case 4:
					// STY - Store Y Register
					cpu.mem.Write(addr, cpu.Y)
				case 5:
					// LDY - Load Y Register
					cpu.Y = cpu.mem.Read(addr)
					cpu.status_Z = cpu.Y == 0
					cpu.status_N = (cpu.Y & 0x80) > 0
				case 6:
					// CPY - Compare Y Register
					data := cpu.mem.Read(addr)
					cpu.status_C = cpu.Y >= data
					cpu.status_Z = cpu.Y == data
					cpu.status_N = (cpu.Y - data) & 0x80 > 0
				case 7:
					// CPX - Compare X Register
					data := cpu.mem.Read(addr)
					cpu.status_C = cpu.X >= data
					cpu.status_Z = cpu.X == data
					cpu.status_N = (cpu.X - data) & 0x80 > 0
				}

				if instructionType != 2 && instructionType != 3 {
					cpu.PC += address(size)
				}
				cycles_left -= cycles
				continue
			}
		}

		if opcode & 0x1F == 0x10 {
			// branch instructions
			var flag bool
			switch (opcode & 0xC0) >> 6 {
			case 0:
				flag = cpu.status_N
			case 1:
				flag = cpu.status_V
			case 2:
				flag = cpu.status_C
			case 3:
				flag = cpu.status_Z
			}
			comp := opcode & 0x20 > 0

			if comp == flag {
				var offset int8 = int8(cpu.mem.Read(cpu.PC + 1))
				destination := address(int32(cpu.PC) + int32(offset)) + 2
				if destination & 0xFF00 != (cpu.PC + 2) & 0xFF00 {
					cycles_left -= 4
				} else {
					cycles_left -= 3
				}
				cpu.PC = destination
			} else {
				cpu.PC += 2
				cycles_left -= 2
			}
			continue
		}

		time.Sleep(100000000)
		panic(fmt.Sprintf("Unknown Opcode at $%.4X $%.2X", cpu.PC, opcode))
	}

	return cycles - cycles_left
}