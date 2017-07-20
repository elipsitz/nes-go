package main

import (
	"bufio"
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"os"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var referenceLogLine string
var referenceLog *bufio.Scanner
var referenceLogBegan bool

func logline(line string) {
	// fmt.Println(line)
	return
	if referenceLog != nil {
		if len(referenceLogLine) == 0 {
			referenceLog.Scan()
			referenceLogLine = referenceLog.Text()
		}
		// fmt.Println(referenceLogLine)
		// fmt.Println(line)
		for i := 0; i < len(line) && i < len(referenceLogLine); i++ {
			if line[i] != referenceLogLine[i] && line[i] != '_' {
				if !referenceLogBegan {
					return
				}
				fmt.Println(referenceLogLine)
				fmt.Println(line)
				time.Sleep(10000000)
				panic("FAIL")
				return
			}
		}
		if !referenceLogBegan {
			fmt.Println("reference log begins")
			referenceLogBegan = true
		}
		fmt.Println(referenceLogLine)
		referenceLogLine = ""
	}
}

var surface *sdl.Surface
var window *sdl.Window
var debugSurface *sdl.Surface
var debugRenderer *sdl.Renderer

var nes *Nes
var debug int
var debugNumScreens int = 2

func sdlInit() {
	var err error
	sdl.Init(sdl.INIT_EVERYTHING)
	window, err = sdl.CreateWindow("aeNES", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 256, 240, sdl.WINDOW_SHOWN)
	check(err)

	surface, err = window.GetSurface()
	check(err)

	debugSurface, err = sdl.CreateRGBSurface(0, 256, 240, 32, 0xff000000, 0x00ff0000, 0x0000ff00, 0x000000ff)
	check(err)

	debugRenderer, err = sdl.CreateSoftwareRenderer(debugSurface)
	check(err)
}

func sdlLoop() {
	var event sdl.Event
	running := true
	for running {
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseMotionEvent:
				// fmt.Printf("[%d ms] MouseMotion\ttype:%d\tid:%d\tx:%d\ty:%d\txrel:%d\tyrel:%d\n", t.Timestamp, t.Type, t.Which, t.X, t.Y, t.XRel, t.YRel)
			case *sdl.KeyDownEvent:
				switch t.Keysym.Scancode {
				case sdl.SCANCODE_RETURN:
					nes.controller1.buttons[ButtonStart] = true
				case sdl.SCANCODE_RSHIFT:
					nes.controller1.buttons[ButtonSelect] = true
				case sdl.SCANCODE_LEFT:
					nes.controller1.buttons[ButtonLeft] = true
				case sdl.SCANCODE_RIGHT:
					nes.controller1.buttons[ButtonRight] = true
				case sdl.SCANCODE_UP:
					nes.controller1.buttons[ButtonUp] = true
				case sdl.SCANCODE_DOWN:
					nes.controller1.buttons[ButtonDown] = true
				case sdl.SCANCODE_Z:
					nes.controller1.buttons[ButtonA] = true
				case sdl.SCANCODE_X:
					nes.controller1.buttons[ButtonB] = true
				}
			case *sdl.KeyUpEvent:
				switch t.Keysym.Scancode {
				case sdl.SCANCODE_RETURN:
					nes.controller1.buttons[ButtonStart] = false
				case sdl.SCANCODE_RSHIFT:
					nes.controller1.buttons[ButtonSelect] = false
				case sdl.SCANCODE_LEFT:
					nes.controller1.buttons[ButtonLeft] = false
				case sdl.SCANCODE_RIGHT:
					nes.controller1.buttons[ButtonRight] = false
				case sdl.SCANCODE_UP:
					nes.controller1.buttons[ButtonUp] = false
				case sdl.SCANCODE_DOWN:
					nes.controller1.buttons[ButtonDown] = false
				case sdl.SCANCODE_Z:
					nes.controller1.buttons[ButtonA] = false
				case sdl.SCANCODE_X:
					nes.controller1.buttons[ButtonB] = false
				case sdl.SCANCODE_GRAVE:
					debug = (debug + 1) % (debugNumScreens + 1)
				case sdl.SCANCODE_TAB:
					for y := 0; y < 30; y++ {
						for x := 0; x < 32; x++ {
							fmt.Printf("%.2X ", nes.ppu.mem.Read(0x2000+address(y*32)+address(x)))
						}
						fmt.Printf("\n")
					}
					sdl.Delay(100000000)
				}
			}
		}

		timeStart := time.Now()
		nes.EmulateFrame()
		timeEnd := time.Now()
		frameTime := timeEnd.Sub(timeStart)
		// desired frame time = 16.66ms = 16666667 nanoseconds
		// fmt.Printf("Frame in: %dms\n", frameTime.Nanoseconds() / 1000000)
		delay := (16666667 - frameTime.Nanoseconds()) / 1000000
		if delay > 0 {
			sdl.Delay(uint32(delay))
		}
	}
}

func pushPixel(x int, y int, col color) {
	pixels := surface.Pixels()
	pixels[4*(y*int(surface.W)+x)+0] = byte(col >> 0)
	pixels[4*(y*int(surface.W)+x)+1] = byte(col >> 8)
	pixels[4*(y*int(surface.W)+x)+2] = byte(col >> 16)
	pixels[4*(y*int(surface.W)+x)+3] = byte(col >> 24)

}

func pushFrame() {
	if debug > 0 {
		if debug == 1 {
			debugRenderer.SetDrawColor(0, 255, 0, 255)
			for i := 0; i < 256; i += 4 {
				x, y := nes.ppu.oam[i+3], nes.ppu.oam[i+0]
				debugRenderer.DrawRect(&sdl.Rect{int32(x), int32(y), 8, 8})
			}
		}

		if debug == 2 {
			// draw pattern tables
			for x := 0; x < 256; x++ {
				for y := 0; y < 128; y++ {
					addr := 0
					addr |= y % 8
					addr |= (x % 128 / 8) << 4
					addr |= (y % 128 / 8) << 8
					if x >= 128 {
						addr |= 0x1000
					}
					lo, hi := nes.ppu.mem.Read(address(addr)), nes.ppu.mem.Read(address(addr+8))
					col := (((lo << uint(x%8)) & 0x80) >> 7) | (((hi << uint(x%8)) & 0x80) >> 6)
					col += 1
					debugRenderer.SetDrawColor(col*60, col*60, col*60, 255)
					debugRenderer.DrawPoint(x, y)
				}
			}
		}

		debugSurface.Blit(nil, surface, nil)
	}

	window.UpdateSurface()
	debugSurface.FillRect(nil, 0x00000000)
}

func sdlCleanup() {
	window.Destroy()
	sdl.Quit()
}

func main() {
	fmt.Println("aeNES")
	// romPath := "roms/Super Mario Bros.nes"
	romPath := "roms/test/test_cpu_exec_space_ppuio.nes"
	fmt.Println("loading", romPath)

	referenceLogFile, err := os.Open(romPath + ".debug")
	if err == nil {
		fmt.Println("found debug log")
		referenceLog = bufio.NewScanner(referenceLogFile)
	}

	nes = NewNes(romPath)
	nes.ppu.funcPushFrame = pushFrame
	nes.ppu.funcPushPixel = pushPixel

	// boot up
	nes.cpu.PC = nes.cpu.getVectorReset()
	fmt.Printf("resetting PC to $%.4X\n", nes.cpu.PC)
	// nes.cpu.PC = 0xC000

	sdlInit()
	sdlLoop()
	sdlCleanup()
}
