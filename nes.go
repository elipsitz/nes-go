package main

type address uint16
type color uint32

type Nes struct {
	cpu       *Cpu
	ppu       *Ppu
	cartridge *Cartridge
	mapper    Mapper

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

func (nes *Nes) Emulate() int {
	clocks := nes.cpu.Emulate(1)
	nes.ppu.Emulate(clocks * 3)

	return clocks
}

func (nes *Nes) EmulateFrame() int {
	cycles := 0
	startFrame := nes.ppu.frameCounter
	for startFrame == nes.ppu.frameCounter {
		cycles += nes.Emulate()
	}
	return cycles
}