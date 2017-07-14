package main

import (
	"fmt"
	"os"
	"io"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	fmt.Println("aeNES")

	f, err := os.Open("roms/donkey-kong.nes")
	check(err)
	r := f // bufio.NewReader(f)

	header := make([]byte, 16)
	_, err = io.ReadFull(r, header)
	check(err)

	fmt.Println(string(header))

	var prg_rom_size, chr_rom_size int = int(header[4]) * 16384, int(header[5]) * 8192;
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

	for i := 0; i < 50000; i++ {
		nes.cpu.Emulate(1)
		nes.ppu.Emulate(3)
	}
}
