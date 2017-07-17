package main

type address uint16
type color uint32

type Nes struct {
	cpu *Cpu
	ppu *Ppu
	cartridge *Cartridge
	mapper Mapper

	ram [2048]byte
}

func NewNes(romPath string) *Nes {
	nes := Nes{
		cartridge: LoadCartridge(romPath),
	}
	nes.cpu = NewCpu(&nes)
	nes.ppu = NewPpu(&nes)
	nes.mapper = NewMapper(&nes)

	return &nes
}