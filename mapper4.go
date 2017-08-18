package main

type MapperMMC3 struct {
	nes *Nes

	irqEnabled bool
	irqLatch   byte
	irqReload  bool

	mirrorMode int // 0: vertical, 1: horizontal

	bankRegisters      [8]int
	bankSelectRegister int
	bankPRGMode        int
	bankCHRMode        int

	prgRam [8192]byte
}

func NewMapperMMC3(nes *Nes) *MapperMMC3 {
	return &MapperMMC3{
		nes:        nes,
		irqEnabled: true,
		irqReload:  false,
	}
}

func (m *MapperMMC3) Read(addr address) byte {
	switch {
	case addr <= 0x1FFF:
		return m.nes.cartridge.chr[m.resolvePpuRomAddr(addr)]
	case addr <= 0x2FFF:
		return m.nes.ppu.vram[TranslateVRamAddress(addr, 1-m.mirrorMode)]
	case addr < 0x6000:
		// ?????
	case addr <= 0x7FFF:
		// internal ram
		return m.prgRam[addr-0x6000]
	case addr <= 0xFFFF:
		return m.nes.cartridge.prg[m.resolveCpuRomAddr(addr)]
	}
	return 0
}

func (m *MapperMMC3) Write(addr address, data byte) {
	switch {
	case addr <= 0x1FFF:
		m.nes.cartridge.chr[m.resolvePpuRomAddr(addr)] = data
	case addr <= 0x2FFF:
		m.nes.ppu.vram[TranslateVRamAddress(addr, 1-m.mirrorMode)] = data
	case addr < 0x6000:
		// ?????
	case addr <= 0x7FFF:
		// write to prg ram
		m.prgRam[addr-0x6000] = data
	case addr <= 0x9FFF && (addr&0x1 == 0):
		// bank select register
		m.bankSelectRegister = int(data & 0x7)
		m.bankPRGMode = int(data&0x40) >> 6
		m.bankCHRMode = int(data&0x80) >> 7
	case addr <= 0x9FFF && (addr&0x1 == 1):
		// TODO bank data register
		if m.bankSelectRegister == 6 || m.bankSelectRegister == 7 {
			data &= 0x3F
		} else if m.bankSelectRegister == 0 || m.bankSelectRegister == 1 {
			data &= 0xFE
		}
		m.bankRegisters[m.bankSelectRegister] = int(data)
	case addr <= 0xBFFF && (addr&0x1 == 0):
		// mirroring register
		m.mirrorMode = int(data & 0x1)
	case addr <= 0xBFFF && (addr&0x1 == 1):
		// TODO PRG RAM protect register
	case addr <= 0xDFFF && (addr&0x1 == 0):
		// IRQ latch register
		m.irqLatch = data
	case addr <= 0xDFFF && (addr&0x1 == 1):
		// IRQ reload register
		m.irqReload = true
	case addr <= 0xFFFF && (addr&0x1 == 0):
		// IRQ disable register
		// TODO acknowledge pending interrupts (??)
		m.irqEnabled = false
	case addr <= 0xFFFF && (addr&0x1 == 1):
		// IRQ enable register
		m.irqEnabled = true
	}
}

func (m *MapperMMC3) resolvePpuRomAddr(addr address) int {
	bank_addr := addr & 0x3FF
	bank_index := int(addr&0x1C00) >> 10

	if m.bankCHRMode != 0 {
		if bank_index >= 4 {
			bank_index -= 4
		} else {
			bank_index += 4
		}
	}

	var bank int
	if bank_index < 4 {
		bank = m.bankRegisters[bank_index/2] | (bank_index & 0x1)
	} else {
		bank = m.bankRegisters[bank_index-2]
	}

	return bank*1024 + int(bank_addr)
}

func (m *MapperMMC3) resolveCpuRomAddr(addr address) int {
	// maps a raw address for the CPU into the ROM (0x8000 to 0xFFFF)
	if m.bankPRGMode == 0 {
		switch {
		case addr <= 0x9FFF:
			return (8192 * m.bankRegisters[6]) + int(addr-0x8000)
		case addr <= 0xBFFF:
			return (8192 * m.bankRegisters[7]) + int(addr-0xA000)
		case addr <= 0xDFFF:
			return (8192*-2 + len(nes.cartridge.prg)) + int(addr-0xC000)
		case addr <= 0xFFFF:
			return (8192*-1 + len(nes.cartridge.prg)) + int(addr-0xE000)
		}
	} else {
		switch {
		case addr <= 0x9FFF:
			return (8192*-2 + len(nes.cartridge.prg)) + int(addr-0x8000)
		case addr <= 0xBFFF:
			return (8192 * m.bankRegisters[7]) + int(addr-0xA000)
		case addr <= 0xDFFF:
			return (8192 * m.bankRegisters[6]) + int(addr-0xC000)
		case addr <= 0xFFFF:
			return (8192*-1 + len(nes.cartridge.prg)) + int(addr-0xE000)
		}
	}

	panic("should be unreachable")
}
