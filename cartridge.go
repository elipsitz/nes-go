package main

import (
	"os"
	"io"
	"hash/crc32"
)

type Cartridge struct {
	header   []byte
	prg      []byte
	chr      []byte
	mapperID int
}

func LoadCartridge(path string) *Cartridge {
	f, err := os.Open(path)
	check(err)
	r := f

	cartridge := Cartridge{}

	cartridge.header = make([]byte, 16)
	_, err = io.ReadFull(r, cartridge.header)
	check(err)

	var prgSize, chrSize int = int(cartridge.header[4])*16384, int(cartridge.header[5])*8192;
	cartridge.prg, cartridge.chr = make([]byte, prgSize), make([]byte, chrSize)
	_, err = io.ReadFull(r, cartridge.prg)
	check(err)
	_, err = io.ReadFull(r, cartridge.chr)
	check(err)
	f.Close()

	cartridge.mapperID = int((cartridge.header[7] & 0xF0) | (cartridge.header[6] >> 4))

	return &cartridge
}

func (cartridge *Cartridge) CRC32() (prg, chr, total uint32) {
	prg, chr = crc32.ChecksumIEEE(cartridge.prg), crc32.ChecksumIEEE(cartridge.chr)
	total = crc32.ChecksumIEEE(append(cartridge.prg, cartridge.chr...))
	return
}
