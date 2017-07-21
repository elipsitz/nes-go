package main

import (
	"bufio"
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"os"
	"reflect"
	"time"
	"unsafe"
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

var window *sdl.Window
var windowRenderer *sdl.Renderer
var windowTexture *sdl.Texture
var buffer [w * h]uint32
var debugSurface *sdl.Surface
var debugRenderer *sdl.Renderer
var debugTexture *sdl.Texture

var nes *Nes
var debug int
var framesRendered int
var fpsTimer time.Time

const debugNumScreens = 2
const scale = 2
const w = 256
const h = 240

var paused bool

func sdlInit() {
	var err error
	sdl.Init(sdl.INIT_EVERYTHING)

	window, windowRenderer, err = sdl.CreateWindowAndRenderer(w*scale, h*scale, 0)
	check(err)
	windowTexture, err = windowRenderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, w, h)
	check(err)

	debugSurface, err = sdl.CreateRGBSurface(0, w*scale, h*scale, 32, 0x00ff0000, 0x0000ff00, 0x000000ff, 0xff000000)
	check(err)
	debugRenderer, err = sdl.CreateSoftwareRenderer(debugSurface)
	check(err)
	debugTexture, err = windowRenderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, w*scale, h*scale)
	debugTexture.SetBlendMode(sdl.BLENDMODE_BLEND)

	debugRenderer.SetScale(scale, scale)
	fpsTimer = time.Now()
}

func sdlLoop() {
	var event sdl.Event
	running := true
	for running {
		frameStart := time.Now()

		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
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
				case sdl.SCANCODE_SPACE:
					paused = !paused
					if paused {
						fmt.Println("Paused.")
					} else {
						fmt.Println("Unpaused.")
					}
				}
			}
		}

		if !paused {
			nes.EmulateFrame()
		}

		frameTime := time.Now().Sub(frameStart)
		delay := (16666667 - frameTime.Nanoseconds()) / 1000000
		if delay > 0 {
			sdl.Delay(uint32(delay))
		}

		framesRendered += 1
		timeSpent := time.Now().Sub(fpsTimer)
		if timeSpent.Seconds() > 1 {
			fps := float64(framesRendered) / timeSpent.Seconds()
			fpsTimer = time.Now()
			framesRendered = 0
			window.SetTitle(fmt.Sprintf("aeNes - FPS: %d", int(fps)))
		}
	}
}

func pushPixel(x int, y int, col color) {
	buffer[y*w+x] = uint32(col)
}

func drawDebug() {
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

		pixels := debugSurface.Pixels()
		hdr := (*reflect.SliceHeader)(unsafe.Pointer(&pixels))
		debugTexture.Update(nil, unsafe.Pointer(hdr.Data), 4*w*scale)
		windowRenderer.Copy(debugTexture, nil, nil)
		debugSurface.FillRect(nil, 0x00000000)
	}
}

func pushFrame() {
	// https://wiki.libsdl.org/MigrationGuide#If_your_game_just_wants_to_get_fully-rendered_frames_to_the_screen
	windowTexture.Update(nil, unsafe.Pointer(&buffer), 4*w)
	windowRenderer.Copy(windowTexture, nil, nil)
	drawDebug()
	windowRenderer.Present()
}

func sdlCleanup() {
	window.Destroy()
	sdl.Quit()
}

func main() {
	fmt.Println("aeNES")
	romPath := "roms/Legend of Zelda, The.nes"
	// romPath := "roms/test/test_cpu_exec_space_ppuio.nes"
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
