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

var referenceLog *bufio.Scanner

func logline(line string) {
	// fmt.Println(line)
	return
	if referenceLog != nil {
		if referenceLog.Scan() {
			reference := referenceLog.Text()
			for i := 0; i < len(line); i++ {
				if line[i] != reference[i] && line[i] != '_' {
					fmt.Println(reference)
					fmt.Println(line)
					time.Sleep(10000000)
					panic("FAIL")
					return
				}
			}
			fmt.Println(reference)
		}
	}
}

var surface *sdl.Surface
var window *sdl.Window
var debugSurface *sdl.Surface
var debugRenderer *sdl.Renderer

var nes *Nes
var debug bool

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
					debug = !debug
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
	if debug {
		debugRenderer.SetDrawColor(0, 255, 0, 255)
		for i := 0; i < 256; i += 4 {
			x, y := nes.ppu.oam[i+3], nes.ppu.oam[i+0]
			debugRenderer.DrawRect(&sdl.Rect{int32(x), int32(y), 8, 8})
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
	romPath := "roms/Metroid.nes"
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
