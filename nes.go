package main

import "fmt"

type address uint16
type color uint32

type Nes struct {
	cpu         *Cpu
	ppu         *Ppu
	cartridge   *Cartridge
	mapper      Mapper
	controller1 *Controller
	controller2 *Controller

	ram [4096]byte // only 2048 bytes are included in the console normally
}

func NewNes(romPath string) *Nes {
	nes := Nes{
		cartridge: LoadCartridge(romPath),
	}
	a, b, c := nes.cartridge.CRC32()
	fmt.Printf("CRC32: %.8X, %.8X, %.8X", a, b, c)
	nes.cpu = NewCpu(&nes)
	nes.ppu = NewPpu(&nes)
	nes.mapper = NewMapper(&nes)
	nes.controller1 = NewController()
	nes.controller2 = NewController()

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
