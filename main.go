package main

import (
	"fmt"
	"bufio"
	"time"
	"github.com/veandco/go-sdl2/sdl"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var referenceLog *bufio.Scanner
func logline(line string) {
	fmt.Println(line)
	return
	if referenceLog.Scan() {
		reference := referenceLog.Text()
		for i := 0; i < len(reference); i++ {
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

var surface *sdl.Surface
var window *sdl.Window

var nes *Nes

func sdlInit() {
	var err error
	sdl.Init(sdl.INIT_EVERYTHING)
	window, err = sdl.CreateWindow("aeNES", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 256, 240, sdl.WINDOW_SHOWN)
	check(err)

	surface, err = window.GetSurface()
	check(err)
}

func sdlLoop() {
	nesLoop()

	var event sdl.Event
	running := true
	for running {
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
				fmt.Println(t.Type)
			case *sdl.MouseMotionEvent:
				// fmt.Printf("[%d ms] MouseMotion\ttype:%d\tid:%d\tx:%d\ty:%d\txrel:%d\tyrel:%d\n", t.Timestamp, t.Type, t.Which, t.X, t.Y, t.XRel, t.YRel)
			}
		}

		nesLoop()
		sdl.Delay(16)
	}
}

func pushPixel(x int, y int, col color) {
	pixels := surface.Pixels()
	pixels[4 * (y * int(surface.W) + x) + 0] = byte(col >> 0)
	pixels[4 * (y * int(surface.W) + x) + 1] = byte(col >> 8)
	pixels[4 * (y * int(surface.W) + x) + 2] = byte(col >> 16)
	pixels[4 * (y * int(surface.W) + x) + 3] = byte(col >> 24)
}

func pushFrame() {
	window.UpdateSurface()
}

func sdlCleanup() {
	window.Destroy()
	sdl.Quit()
}

func nesLoop() {
	// NES clock rate
	clock := 1789773
	for i := 0; i < clock / 60; i++ {
		nes.cpu.Emulate(1)
		nes.ppu.Emulate(3)
	}
}

func main() {
	fmt.Println("aeNES")

	nes = NewNes("roms/Donkey Kong.nes")
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
