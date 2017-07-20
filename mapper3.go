package main

type Mapper3 struct {
	nes  *Nes
	bank int
}

func NewMapper3(nes *Nes) *Mapper3 {
	return &Mapper3{
		nes: nes,
	}
}

func (m *Mapper3) Read(addr address) byte {
	switch {
	case addr <= 0x1FFF:
		return m.nes.cartridge.chr[address(m.bank*8192)+addr]
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

func (m *Mapper3) Write(addr address, data byte) {
	if addr >= 0x8000 && addr <= 0xFFFF {
		m.bank = int(data) % (len(m.nes.cartridge.chr) / 8192)
	}
}
