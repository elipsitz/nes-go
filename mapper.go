package main

import "fmt"

type Mapper interface {
	Read(addr address) byte
	Write(addr address, data byte)
}

func NewMapper(nes *Nes) Mapper {
	switch nes.cartridge.mapperID {
	case 0:
		return &MapperMMC0{}
	default:
		panic(fmt.Sprintf("Unknown mapper: %d", nes.cartridge.mapperID))
	}
}

type MapperMMC0 struct {
}

func (m *MapperMMC0) Read(addr address) byte {
	switch {
	case addr <= 0x1FFF:
		return nes.cartridge.chr[addr]
	case addr <= 0x2FFF:
		// mirroring of nametables.
		// TODO actually implement
		return nes.ppu.vram[(addr-0x2000)&0x7FF]
	case addr >= 0x8000 && addr <= 0xBFFF:
		return nes.cartridge.prg[addr-0x8000]
	case addr >= 0xC000 && addr <= 0xFFFF:
		if len(nes.cartridge.prg) > 0x4000 {
			return nes.cartridge.prg[addr-0x8000]
		} else {
			return nes.cartridge.prg[addr-0xC000]
		}
	default:
		panic(fmt.Sprintf("MMC read out of bounds: %.4X", addr))
	}
}

func (m *MapperMMC0) Write(addr address, data byte) {
	// TODO
	switch {
	case addr <= 0x1FFF:
		nes.cartridge.chr[addr] = data // Not really ROm though
	case addr <= 0x2FFF:
		// TODO actually implement nametable mirroring
		nes.ppu.vram[(addr-0x2000)&0x7FF] = data
	default:
		panic(fmt.Sprintf("MMC write out of bounds: %.4X", addr))
	}
}
