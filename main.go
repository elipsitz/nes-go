package main

import (
	"fmt"
	"os"
	"io"
	"bufio"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var referenceLog *bufio.Scanner
func logline(line string) {
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

func main() {
	fmt.Println("aeNES")

	f, err := os.Open("roms/nestest.nes")
	check(err)
	r := f

	logFile, _ := os.Open("roms/nestest.log")
	defer logFile.Close()
	referenceLog = bufio.NewScanner(logFile)

	header := make([]byte, 16)
	_, err = io.ReadFull(r, header)
	check(err)

	var prg_rom_size, chr_rom_size int = int(header[4]) * 16384, int(header[5]) * 8192;
	fmt.Printf("PRG size: %d, CHR size: %d\n", prg_rom_size, chr_rom_size)
	prg_rom, chr_rom := make([]byte, prg_rom_size), make([]byte, chr_rom_size)
	_, err = io.ReadFull(r, prg_rom)
	check(err)
	_, err = io.ReadFull(r, chr_rom)
	check(err)
	f.Close()

	nes := Nes{
		prg_rom: prg_rom,
		chr_rom: chr_rom,
	}
	nes.cpu = NewCpu(&nes)

	// boot up
	nes.cpu.PC = nes.getVectorReset()
	fmt.Println("resetting PC to", nes.cpu.PC)
	nes.cpu.PC = 0xC000

	for i := 0; i < 10000; i++ {
		nes.cpu.Emulate(1)
		nes.ppu.Emulate(3)
	}
}
