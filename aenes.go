package main

import (
	"fmt"
	"os"
	"bufio"
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
	r := bufio.NewReader(f)

	header := make([]byte, 16)
	_, err = r.Read(header)
	check(err)

	fmt.Println(string(header))

	var prg_rom_size, chr_rom_size int = int(header[4]) * 16384, int(header[5]) * 8192;
	prg_rom, chr_rom := make([]byte, prg_rom_size), make([]byte, chr_rom_size)
	_, _ = r.Read(prg_rom)
	_, _ = r.Read(chr_rom)

	f.Close()


}
