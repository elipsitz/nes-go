package main

import (
	"fmt"
)

type Ppu struct {
	nes *Nes
	mem Memory

	// drawing interfaces
	funcPushPixel func(int, int, color)
	funcPushFrame func()

	vram          [2048]byte
	oam           [256]byte
	secondary_oam [32]byte
	palette       [32]byte
	colors        [64]color

	warmupRemaining int
	scanlineCounter int
	tickCounter     int
	frameCounter    int
	cycles          uint64

	status_rendering    bool
	flag_vBlank         byte
	flag_sprite0Hit     byte
	flag_spriteOverflow byte

	ppuLatch          byte
	addressLatch      uint16
	addressLatchIndex byte

	// sprite rendering
	spriteEvaluationN         int
	spriteEvaluationM         int
	spriteEvaluationRead      byte
	pendingNumScanlineSprites int
	numScanlineSprites        int
	spriteXCounters           [8]int
	spriteAttributes          [8]byte
	spriteBitmapDataLo        [8]byte
	spriteBitmapDataHi        [8]byte

	// PPUCTRL
	flag_baseNametable          byte
	flag_incrementVram          byte
	flag_spriteTableAddress     byte
	flag_backgroundTableAddress byte
	flag_spriteSize             byte
	flag_masterSlave            byte
	flag_generateNMIs           byte

	// PPUMASK
	flag_grayscale          byte
	flag_showSpritesLeft    byte
	flag_showBackgroundLeft byte
	flag_renderSprites      byte
	flag_renderBackground   byte
	flag_emphasizeRed       byte
	flag_emphasizeGreen     byte
	flag_emphasizeBlue      byte

	scrollPositionX byte
	scrollPositionY byte

	oamAddr byte
}

func NewPpu(nes *Nes) *Ppu {
	fmt.Println("...")
	return &Ppu{
		nes:             nes,
		mem:             &PPUMemory{nes: nes},
		warmupRemaining: 29658 * 3,
		scanlineCounter: -1, // counts scanlines in a frame ( https://wiki.nesdev.com/w/index.php/PPU_rendering#Line-by-line_timing )
		tickCounter:     0,  // counts clock cycle ticks in a scanline
		frameCounter:    0,  // counts total frames (vblanks)

		flag_vBlank: 0,
		colors:      [64]color{84*256*256 + 84*256 + 84, 0*256*256 + 30*256 + 116, 8*256*256 + 16*256 + 144, 48*256*256 + 0*256 + 136, 68*256*256 + 0*256 + 100, 92*256*256 + 0*256 + 48, 84*256*256 + 4*256 + 0, 60*256*256 + 24*256 + 0, 32*256*256 + 42*256 + 0, 8*256*256 + 58*256 + 0, 0*256*256 + 64*256 + 0, 0*256*256 + 60*256 + 0, 0*256*256 + 50*256 + 60, 0*256*256 + 0*256 + 0, 0*256*256 + 0*256 + 0, 0*256*256 + 0*256 + 0, 152*256*256 + 150*256 + 152, 8*256*256 + 76*256 + 196, 48*256*256 + 50*256 + 236, 92*256*256 + 30*256 + 228, 136*256*256 + 20*256 + 176, 160*256*256 + 20*256 + 100, 152*256*256 + 34*256 + 32, 120*256*256 + 60*256 + 0, 84*256*256 + 90*256 + 0, 40*256*256 + 114*256 + 0, 8*256*256 + 124*256 + 0, 0*256*256 + 118*256 + 40, 0*256*256 + 102*256 + 120, 0*256*256 + 0*256 + 0, 0*256*256 + 0*256 + 0, 0*256*256 + 0*256 + 0, 236*256*256 + 238*256 + 236, 76*256*256 + 154*256 + 236, 120*256*256 + 124*256 + 236, 176*256*256 + 98*256 + 236, 228*256*256 + 84*256 + 236, 236*256*256 + 88*256 + 180, 236*256*256 + 106*256 + 100, 212*256*256 + 136*256 + 32, 160*256*256 + 170*256 + 0, 116*256*256 + 196*256 + 0, 76*256*256 + 208*256 + 32, 56*256*256 + 204*256 + 108, 56*256*256 + 180*256 + 204, 60*256*256 + 60*256 + 60, 0*256*256 + 0*256 + 0, 0*256*256 + 0*256 + 0, 236*256*256 + 238*256 + 236, 168*256*256 + 204*256 + 236, 188*256*256 + 188*256 + 236, 212*256*256 + 178*256 + 236, 236*256*256 + 174*256 + 236, 236*256*256 + 174*256 + 212, 236*256*256 + 180*256 + 176, 228*256*256 + 196*256 + 144, 204*256*256 + 210*256 + 120, 180*256*256 + 222*256 + 120, 168*256*256 + 226*256 + 144, 152*256*256 + 226*256 + 180, 160*256*256 + 214*256 + 228, 160*256*256 + 162*256 + 160, 0*256*256 + 0*256 + 0, 0*256*256 + 0*256 + 0},
	}
}

func (ppu *Ppu) ReadRegister(register int) byte {
	switch register {
	case 2:
		// PPUSTATUS
		var status byte = ppu.ppuLatch & 0x1F
		status |= ppu.flag_spriteOverflow << 5
		status |= ppu.flag_sprite0Hit << 6
		status |= ppu.flag_vBlank << 7

		ppu.flag_vBlank = 0
		ppu.ppuLatch = status
		ppu.addressLatchIndex = 0
		return status
	case 4:
		// OAMDATA
		// TODO if visible scanline and cycle between 1-64, return 0xFF
		return ppu.oam[ppu.oamAddr]
		// XXX increment after read during rendering?
	case 7:
		// PPUDATA
		data := ppu.mem.Read(address(ppu.addressLatch))
		if ppu.flag_incrementVram == 0 {
			ppu.addressLatch += 1
		} else {
			ppu.addressLatch += 32
		}
		return data
	default:
		return ppu.ppuLatch
	}
}

func (ppu *Ppu) WriteRegister(register int, data byte) {
	ppu.ppuLatch = data
	switch register {
	case 0:
		// PPUCTRL
		if ppu.cycles > 29658*3 {
			ppu.flag_baseNametable = data & 0x3
			ppu.flag_incrementVram = data & 0x4 >> 2
			ppu.flag_spriteTableAddress = data & 0x8 >> 3
			ppu.flag_backgroundTableAddress = data & 0x10 >> 4
			ppu.flag_spriteSize = data & 0x20 >> 5
			ppu.flag_masterSlave = data & 0x40 >> 6
			ppu.flag_generateNMIs = data & 0x80 >> 7
		}
	case 1:
		// PPUMASK
		ppu.flag_grayscale = data & 0x1 >> 0
		ppu.flag_showBackgroundLeft = data & 0x2 >> 1
		ppu.flag_showSpritesLeft = data & 0x4 >> 2
		ppu.flag_renderBackground = data & 0x8 >> 3
		ppu.flag_renderSprites = data & 0x10 >> 4
		ppu.flag_emphasizeRed = data & 0x20 >> 5
		ppu.flag_emphasizeGreen = data & 0x40 >> 6
		ppu.flag_emphasizeBlue = data & 0x80 >> 7
	case 3:
		// OAMADDR
		ppu.oamAddr = data
	case 4:
		// OAMDATA
		if !ppu.status_rendering {
			ppu.oam[ppu.oamAddr] = data
			ppu.oamAddr++
		}
	case 5:
		// PPUSCROLL
		if ppu.addressLatchIndex == 0 {
			ppu.scrollPositionX = data
		} else {
			ppu.scrollPositionY = data
		}
		ppu.addressLatchIndex = 1 - ppu.addressLatchIndex
	case 6:
		// PPUADDR
		if ppu.addressLatchIndex == 0 {
			ppu.addressLatch = (ppu.addressLatch & 0x00FF) | (uint16(data) << 8)
		} else {
			ppu.addressLatch = (ppu.addressLatch & 0xFF00) | (uint16(data))
		}
		// fmt.Println("write to address latch: ", ppu.addressLatchIndex, data, ppu.addressLatch)
		ppu.addressLatchIndex = 1 - ppu.addressLatchIndex
	case 7:
		// PPUDATA
		// fmt.Println("write to ppudata ", ppu.addressLatch, data)
		ppu.mem.Write(address(ppu.addressLatch), data)
		if ppu.flag_incrementVram == 0 {
			ppu.addressLatch += 1
		} else {
			ppu.addressLatch += 32
		}
	case 0x4014:
		// OAMDMA
		// TODO suspend CPU for 513-514 cycles
		addr := address(data) << 8
		for i := 0; i < 256; i++ {
			addr2 := addr + address(i)
			data := nes.cpu.mem.Read(addr2)
			ppu.oam[(ppu.oamAddr+byte(i))&0xFF] = data
		}
	}
}

func (ppu *Ppu) Emulate(cycles int) {
	cycles_left := cycles
	for cycles_left > 0 {
		ppu.cycles++
		ppu.tickCounter++
		if ppu.tickCounter == 341 || (ppu.tickCounter == 340 && ppu.scanlineCounter == -1 && ppu.frameCounter%2 == 1) {
			ppu.tickCounter = 0
			ppu.scanlineCounter++
			if ppu.scanlineCounter > 260 {
				ppu.scanlineCounter = -1
			}
		}

		if ppu.scanlineCounter == -1 {
			if ppu.tickCounter == 0 {
				if ppu.frameCounter%2 == 1 {
					ppu.tickCounter++
				}
			}
			if ppu.tickCounter == 1 {
				// prerender
				ppu.flag_sprite0Hit = 0
				ppu.flag_vBlank = 0
				ppu.flag_spriteOverflow = 1
				ppu.status_rendering = true
			}
		}

		if ppu.scanlineCounter >= 0 && ppu.scanlineCounter < 240 {
			if ppu.tickCounter >= 1 && ppu.tickCounter <= 64 {
				// https://wiki.nesdev.com/w/index.php/PPU_sprite_evaluation
				// Sprite Evaluation Stage 1: Clearing the Secondary OAM
				if ppu.tickCounter%2 == 0 {
					ppu.secondary_oam[(ppu.tickCounter-1)/2] = 0xFF
				}
			}
			if ppu.tickCounter == 65 {
				ppu.spriteEvaluationN = 0
				ppu.spriteEvaluationM = 0
				ppu.pendingNumScanlineSprites = 0
			}
			if ppu.tickCounter >= 65 && ppu.tickCounter <= 256 {
				// Sprite Evaluation Stage 2: Loading the Secondary OAM
				if ppu.spriteEvaluationN < 64 && ppu.pendingNumScanlineSprites < 8 {
					if ppu.tickCounter%2 == 1 {
						// read from primary
						ppu.spriteEvaluationRead = ppu.oam[4*ppu.spriteEvaluationN+ppu.spriteEvaluationM]
					} else {
						// write to secondary
						ppu.secondary_oam[4*ppu.pendingNumScanlineSprites+ppu.spriteEvaluationM] = ppu.spriteEvaluationRead
						if ppu.spriteEvaluationM == 0 {
							// check to see if it's in range
							if byte(ppu.scanlineCounter) >= ppu.spriteEvaluationRead && byte(ppu.scanlineCounter) < ppu.spriteEvaluationRead+8 {
								// it's in range!
							} else {
								// not in range.
								ppu.spriteEvaluationM--
								ppu.spriteEvaluationN++
							}
						}
						if ppu.spriteEvaluationM == 3 {
							ppu.spriteEvaluationN++
							ppu.spriteEvaluationM = 0
							ppu.pendingNumScanlineSprites += 1
						} else {
							ppu.spriteEvaluationM++
						}
					}
				}
			}
			if ppu.tickCounter >= 257 && ppu.tickCounter <= 320 {
				ppu.spriteEvaluationN = (ppu.tickCounter - 257) / 8
				ppu.numScanlineSprites = ppu.pendingNumScanlineSprites
				if (ppu.tickCounter-257)%8 == 0 {
					// fetch x position, attribute into temporary latches and counters
					var ypos, tile, attribute, xpos byte
					if ppu.spriteEvaluationN < ppu.numScanlineSprites {
						ypos = ppu.secondary_oam[ppu.spriteEvaluationN*4+0]
						tile = ppu.secondary_oam[ppu.spriteEvaluationN*4+1]
						attribute = ppu.secondary_oam[ppu.spriteEvaluationN*4+2]
						xpos = ppu.secondary_oam[ppu.spriteEvaluationN*4+3]
					} else {
						ypos, tile, attribute, xpos = 0xFF, 0xFF, 0xFF, 0xFF
					}
					ppu.spriteXCounters[ppu.spriteEvaluationN], ppu.spriteAttributes[ppu.spriteEvaluationN] = int(xpos), attribute

					// TODO support 8x16 sprites
					// fetch bitmap data into shift registers
					tileRow := ppu.scanlineCounter - int(ypos)
					if attribute&0x80 > 0 {
						// flip sprite vertically
						tileRow = 7 - tileRow
					}
					var patternAddr address = 0
					patternAddr |= address(tileRow)
					patternAddr |= address(tile) << 4
					patternAddr |= address(ppu.flag_spriteTableAddress) << 12
					lo, hi := ppu.mem.Read(patternAddr), ppu.mem.Read(patternAddr+8)

					if attribute&0x40 > 0 {
						// flip sprite horizontally
						var hi2, lo2 byte
						for i := 0; i < 8; i++ {
							hi2 = (hi2 << 1) | (hi & 1)
							lo2 = (lo2 << 1) | (lo & 1)
							hi >>= 1
							lo >>= 1
						}
						lo, hi = lo2, hi2
					}

					ppu.spriteBitmapDataLo[ppu.spriteEvaluationN] = lo
					ppu.spriteBitmapDataHi[ppu.spriteEvaluationN] = hi
					// TODO just load transparent sprite if there are fewer than 8
				}
			}

			// drawing!
			if ppu.tickCounter >= 1 && ppu.tickCounter <= 256 {
				// TODO actual scrolling
				tileX, tileY := ppu.tickCounter/8, ppu.scanlineCounter/8
				nametableEntry := ppu.mem.Read(address(0x2000 + tileY*32 + tileX))
				attributeEntry := ppu.mem.Read(address(0x23C0 + (tileY / 4 * 8) + (tileX / 4)))

				var patternTableAddressLo address = 0
				patternTableAddressLo |= address(ppu.scanlineCounter) % 8
				patternTableAddressLo |= address(nametableEntry) << 4
				patternTableAddressLo |= address(ppu.flag_backgroundTableAddress) << 12
				patternTableAddressHi := patternTableAddressLo | 0x8

				color := (ppu.mem.Read(patternTableAddressLo) >> byte(7-(ppu.tickCounter%8)) & 1) | ((ppu.mem.Read(patternTableAddressHi) >> byte(7-(ppu.tickCounter%8)) & 1) << 1)

				palette := (attributeEntry >> byte(((tileX%4)/2)*2+((tileY%4)/2)*4)) & 0x3
				paletteEntry := color + (4 * palette)
				if color == 0 {
					paletteEntry = 0
				}

				// TODO sprite 0 hit
				// check on sprites

				var spritePaletteEntry byte = 0
				for n := ppu.numScanlineSprites - 1; n >= 0; n-- {
					if ppu.spriteXCounters[n] > -7 {
						ppu.spriteXCounters[n]--
						if ppu.spriteXCounters[n] <= 0 {
							// draw this sprite
							attributes := ppu.spriteAttributes[n]
							data := ((ppu.spriteBitmapDataHi[n] & 0x80) >> 6) | ((ppu.spriteBitmapDataLo[n] & 0x80) >> 7)
							ppu.spriteBitmapDataHi[n] <<= 1
							ppu.spriteBitmapDataLo[n] <<= 1
							if data != 0 {
								spritePaletteEntry = 0x10 + data + 4*(attributes&0x3)
							}
						}
					}
				}

				if spritePaletteEntry != 0 {
					// TODO actual priority calculation
					paletteEntry = spritePaletteEntry
				}
				ppu.funcPushPixel(ppu.tickCounter-1, ppu.scanlineCounter, ppu.FetchColor(paletteEntry))
			}
		}

		if ppu.scanlineCounter == 241 && ppu.tickCounter == 1 {
			// VBLANK
			ppu.funcPushFrame()
			if ppu.flag_generateNMIs == 1 {
				ppu.nes.cpu.pendingNmiInterrupt = true
			}
			ppu.flag_vBlank = 1
			ppu.frameCounter += 1
			ppu.status_rendering = false
		}

		cycles_left--
	}
}

func (ppu *Ppu) FetchColor(index byte) color {
	return ppu.colors[ppu.palette[index&0x1F]]
}
