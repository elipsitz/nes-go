package main

type MapperMMC1 struct {
	nes *Nes

	shiftRegister byte
	shiftNumber   int

	mirrorMode int

	registerControl byte
	registerCHR0    byte
	registerCHR1    byte
	registerPRG     byte

	prgRam [8192]byte
}

func (m *MapperMMC1) Read(addr address) byte {
	switch {
	case addr <= 0x0FFF:
		// CHR bank 1
		return m.nes.cartridge.chr[m.getCHR1Index(addr)]
	case addr <= 0x1FFF:
		// CHR bank 2
		return m.nes.cartridge.chr[m.getCHR2Index(addr)]
	case addr <= 0x2FFF:
		// mirroring
		return m.nes.ppu.vram[TranslateVRamAddress(addr, m.mirrorMode)]
	case addr < 0x6000:
		// ?????
		return 0
	case addr >= 0x6000 && addr <= 0x7FFF:
		// internal ram
		return m.prgRam[addr-0x6000]
	case addr <= 0xBFFF:
		// PRG bank 1
		switch (m.registerControl & 0xC) >> 2 {
		case 0:
			fallthrough
		case 1:
			// switch 32 KB at $8000, ignoring low bit of bank number
			return m.nes.cartridge.prg[16384*int(m.registerPRG&0xFE)+int(addr-0x8000)]
		case 2:
			// fix first bank at $8000
			return m.nes.cartridge.prg[int(addr-0x8000)]
		case 3:
			// switch 16 KB bank at $8000
			return m.nes.cartridge.prg[16384*int(m.registerPRG)+int(addr-0x8000)]
		}
	case addr <= 0xFFFF:
		// PRG bank 2
		switch (m.registerControl & 0xC) >> 2 {
		case 0:
			fallthrough
		case 1:
			return m.nes.cartridge.prg[16384*int(m.registerPRG|0x1)+int(addr-0xC000)]
		case 2:
			// switch 16 KB bank at $C000
			return m.nes.cartridge.prg[16384*int(m.registerPRG)+int(addr-0xC000)]
		case 3:
			// fix last bank at $C000
			return m.nes.cartridge.prg[len(m.nes.cartridge.prg)-16384+int(addr-0xC000)]
		}
	}
	return 0
}

func (m *MapperMMC1) getCHR1Index(addr address) int {
	bank := int(m.registerCHR0)
	if m.registerControl&0x10 == 0 {
		// 8KB mode
		bank &= 0xFE
	}
	bank %= len(m.nes.cartridge.chr) / 4096 // XXX is this correct behavior?
	return int(addr-0x0000) + (bank * 4096)
}

func (m *MapperMMC1) getCHR2Index(addr address) int {
	var bank int
	if m.registerControl&0x10 == 0 {
		// 8 KB mode
		bank = int(m.registerCHR0) | 0x1
	} else {
		bank = int(m.registerCHR1)
	}
	bank %= len(m.nes.cartridge.chr) / 4096 // XXX is this correct behavior?
	return int(addr-0x1000) + (bank * 4096)
}

func (m *MapperMMC1) Write(addr address, data byte) {
	if addr < 0x6000 {
		if addr <= 0x0FFF {
			m.nes.cartridge.chr[m.getCHR1Index(addr)] = data
		} else if addr <= 0x1FFF {
			m.nes.cartridge.chr[m.getCHR2Index(addr)] = data
		} else if addr <= 0x2FFF {
			m.nes.ppu.vram[TranslateVRamAddress(addr, m.mirrorMode)] = data
		}
	} else if addr <= 0x7FFF {
		if m.registerPRG&0x10 == 0 {
			m.prgRam[addr-0x6000] = data
		}
	} else {
		// TODO ignore writes on consecutive cycles
		if data&0x80 > 0 {
			// clear shift register
			m.shiftNumber = 0
			m.shiftRegister = 0
		} else {
			// add to shift register
			m.shiftRegister = m.shiftRegister | ((data & 0x1) << uint(m.shiftNumber))
			m.shiftNumber++
		}

		if m.shiftNumber == 5 {
			switch (addr >> 13) & 0x3 {
			case 0:
				m.registerControl = m.shiftRegister

				switch m.registerControl & 0x3 {
				case 0:
					m.mirrorMode = MirrorSingleA
				case 1:
					m.mirrorMode = MirrorSingleB
				case 2:
					m.mirrorMode = MirrorVertical
				case 3:
					m.mirrorMode = MirrorHorizontal
				}
			case 1:
				m.registerCHR0 = m.shiftRegister
			case 2:
				m.registerCHR1 = m.shiftRegister
			case 3:
				m.registerPRG = m.shiftRegister
			}
			m.shiftNumber = 0
			m.shiftRegister = 0
		}
	}
}
