package main

import "fmt"

type Nes struct {
	cpu Cpu6502

	ram [2048]byte
	prg_rom []byte
	chr_rom []byte
}

func (nes *Nes) read_byte(addr address) byte {
	// see https://wiki.nesdev.com/w/index.php/CPU_memory_map

	if addr <= 0x1FFF {
		// internal ram
		return nes.ram[addr & 0x0800]
	} else if addr <= 0x2007 {
		// TODO PPU registers
		fmt.Printf("reading from PPU register $%x\n", addr)
		return 0
	} else if addr <= 0x4017 {
		// TODO NES APU and I/O registers
		return 0
	} else if addr <= 0x401F {
		// CPU test mode (nothing goes here atm)
		return 0
	}

	// OTHERWISE MAPPER
	// TODO mapper
	// XXX currently hardcoded NROM

	if addr >= 0x8000 && addr <= 0xBFFF {
		return nes.prg_rom[addr - 0x8000]
	}
	if addr >= 0xC000 && addr <= 0xFFFF {
		return nes.prg_rom[addr - 0xC000]
	}

	return 0; // shouldn't reach this point
}

func (nes *Nes) write_byte(addr address, data byte) {
	if addr <= 0x1FFF {
		nes.ram[addr & 0x0800] = data;
	} else if addr <= 0x2007 {
		// TODO PPU registers
		fmt.Printf("writing $%x to PPU register $%x\n", data, addr)
	}
	// TODO complete
}

func (nes *Nes) read_uint16(addr address) uint16 {
	return uint16(nes.read_byte(addr)) | (uint16(nes.read_byte(addr + 1)) << 8)
}

func (nes *Nes) getVectorReset() address {
	return address(nes.read_uint16(0xFFFC))
}

func (nes *Nes) getVectorNMI() address {
	return address(nes.read_uint16(0xFFFA))
}

func (nes *Nes) getVectorBRK() address {
	return address(nes.read_uint16(0xFFFE))
}