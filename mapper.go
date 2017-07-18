package main

import "fmt"

type Mapper interface {
	Read(addr address) byte
	Write(addr address, data byte)
}

func NewMapper(nes *Nes) Mapper {
	switch nes.cartridge.mapperID {
	case 0:
		return &MapperMMC0{nes}
	default:
		panic(fmt.Sprintf("Unknown mapper: %d", nes.cartridge.mapperID))
	}
}

const (
	MirrorHorizontal = 0
	MirrorVertical   = 1
	MirrorSingleA    = 2
	MirrorSingleB    = 3
	MirrorFour       = 4
)

var MirrorLookup = [][4]int{
	{0, 0, 1, 1},
	{0, 1, 0, 1},
	{0, 0, 0, 0},
	{1, 1, 1, 1},
	{0, 1, 2, 3},
}

// from nametable address in range 0x2000 to 0x2FFF to VRAM (0x0000 to 0x1000)
func TranslateVRamAddress(addr address, mirrorMode int) int {
	addr -= 0x2000
	bank := MirrorLookup[mirrorMode][addr/(0x400)]
	return (bank * 0x400) + int(addr%0x400)
}

type MapperMMC0 struct {
	nes *Nes
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
		panic(fmt.Sprintf("MMC read out of bounds: %.4X", addr))
	}
}

func (m *MapperMMC0) Write(addr address, data byte) {
	// TODO
	switch {
	case addr <= 0x1FFF:
		m.nes.cartridge.chr[addr] = data // if RAM
	case addr <= 0x2FFF:
		// TODO actually implement nametable mirroring
		m.nes.ppu.vram[(addr-0x2000)&0x7FF] = data
	default:
		panic(fmt.Sprintf("MMC write out of bounds: %.4X", addr))
	}
}
