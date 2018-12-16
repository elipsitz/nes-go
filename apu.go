package main

import "fmt"

type Apu struct {
	nes *Nes
}

func NewApu(nes *Nes) *Apu {
	return &Apu{
		nes: nes,
	}
}

func (apu *Apu) ReadRegister(register int) byte {
	fmt.Printf("apu read from %d\n", register)
	return 0
}

func (apu *Apu) WriteRegister(register int, data byte) {
	fmt.Printf("apu write to %d val %d\n", register, data)
}
