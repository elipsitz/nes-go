package main

type MapperMMC0 struct {
	nes *Nes
}

func NewMapperMMC0(nes *Nes) *MapperMMC0 {
	return &MapperMMC0{
		nes: nes,
	}
}

func (m *MapperMMC0) Read(addr address) byte {
	switch {
	case addr <= 0x1FFF:
		return m.nes.cartridge.chr[addr]
	case addr <= 0x2FFF:
		return m.nes.ppu.vram[TranslateVRamAddress(addr, m.nes.cartridge.mirrorMode)]
	case addr >= 0x8000 && addr <= 0xBFFF:
		return m.nes.cartridge.prg[addr-0x8000]
	case addr >= 0xC000 && addr <= 0xFFFF:
		if len(m.nes.cartridge.prg) > 0x4000 {
			return m.nes.cartridge.prg[addr-0x8000]
		} else {
			return m.nes.cartridge.prg[addr-0xC000]
		}
	default:
		//panic(fmt.Sprintf("MMC read out of bounds: %.4X", addr))
	}
	return 0
}

func (m *MapperMMC0) Write(addr address, data byte) {
	switch {
	case addr <= 0x1FFF:
		m.nes.cartridge.chr[addr] = data // if RAM
	case addr <= 0x2FFF:
		m.nes.ppu.vram[TranslateVRamAddress(addr, m.nes.cartridge.mirrorMode)] = data
	default:
		// panic(fmt.Sprintf("MMC write out of bounds: %.4X", addr))
	}
}
