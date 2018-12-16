package main

type Apu struct {
	nes *Nes
}

func NewApu(nes *Nes) *Apu {
	return &Apu{
		nes:      nes,
	}
}