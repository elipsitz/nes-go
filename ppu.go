package main

type Ppu struct {
	nes *Nes

	warmupRemaining int
	scanlineCounter int
	tickCounter     int

	status_vBlank bool
	status_sprite0Hit bool
	status_spriteOverflow bool
}

func NewPpu(nes *Nes) Ppu {
	return Ppu{
		nes:             nes,
		warmupRemaining: 29658 * 3,
		scanlineCounter: -1, // counts scanlines in a frame ( https://wiki.nesdev.com/w/index.php/PPU_rendering#Line-by-line_timing )
		tickCounter:     0,  // counts clock cycle ticks in a scanline

		status_vBlank: false, // XXX maybe make flags
	}
}

func (ppu *Ppu) ReadRegister (register int) byte {
	if register == 2 {
		// PPUSTATUS TODO do this correctly

		var status byte = 0
		if ppu.status_vBlank {
			status |= 1 << 7;
		}
		ppu.status_vBlank = false
		return status
	}
	return 0;
}

func (ppu *Ppu) WriteRegister (register int, data byte) {
	if register == 0 {
		// fmt.Printf("%X , %b\n", data, data)
	}
}

func (ppu *Ppu) Emulate(cycles int) {
	// do nothing during warmup cycles
	/* if ppu.warmupRemaining > 0 {
		if cycles > ppu.warmupRemaining {
			cycles -= ppu.warmupRemaining;
			ppu.warmupRemaining = 0;
		} else {
			ppu.warmupRemaining -= cycles;
			return;
		}
	}*/

	cycles_left := cycles
	for cycles_left > 0 {
		ppu.tickCounter++
		if ppu.tickCounter >= 341 {
			ppu.tickCounter = 0
			ppu.scanlineCounter++;
			if ppu.scanlineCounter > 260 {
				ppu.scanlineCounter = -1;
			}
		}

		if ppu.scanlineCounter == -1 && ppu.tickCounter == 1 {
			// prerender
			ppu.status_sprite0Hit = false
			ppu.status_vBlank = false
			ppu.status_spriteOverflow = false
		}

		if ppu.scanlineCounter == 241 && ppu.tickCounter == 1 {
			// VBLANK
			ppu.nes.cpu.pendingNmiInterrupt = true
			ppu.status_vBlank = true
		}

		cycles_left--;
	}
}