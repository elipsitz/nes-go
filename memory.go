package main

type Memory interface {
	Read(addr address) byte
	Write(addr address, data byte)
}

func ReadUint16(m Memory, addr address) uint16 {
	return uint16(m.Read(addr)) | (uint16(m.Read(addr + 1)) << 8)
}

type CPUMemory struct {
	nes *Nes
}

func (*CPUMemory) Read(addr address) byte {
	// see https://wiki.nesdev.com/w/index.php/CPU_memory_map
	switch {
	case addr <= 0x1FFF:
		return nes.ram[addr & 0x07FF]
	case addr <= 0x3FFF:
		return nes.ppu.ReadRegister(int(addr & 0x7))
	case addr <= 0x4017:
		// TODO NES APU and I/O REGISTERS
		return 0
	case addr <= 0x401F:
		// CPU test mode
		return 0
	default:
		return nes.mapper.Read(addr)
	}
}

func (*CPUMemory) Write(addr address, data byte) {
	// fmt.Println("mem write", addr, data)

	switch {
	case addr <= 0x1FFF:
		nes.ram[addr & 0x07FF] = data
	case addr <= 0x3FFF:
		nes.ppu.WriteRegister(int(addr & 0x7), data)
	}
	// TODO complete
}

type PPUMemory struct {
	nes *Nes
}