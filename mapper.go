package main

import "fmt"

type Mapper interface {
	Read(addr address) byte
	Write(addr address, data byte)
}

func NewMapper(nes *Nes) Mapper {
	switch nes.cartridge.mapperID {
	case 0:
		return &MapperMMC0{nes: nes}
	case 1:
		return &MapperMMC1{nes: nes}
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
