package main

import (
	"encoding/binary"
	"hash/crc32"
	"io"
	"os"
)

type INesHeader struct {
	MagicNumber uint32
	SizeRomPRG  byte
	SizeRomCHR  byte
	Flag6       byte
	Flag7       byte
	SizeRamPRG  byte
	ExtraFlags  [7]byte
}

type Cartridge struct {
	header     INesHeader
	prg        []byte
	chr        []byte
	mapperID   int
	mirrorMode int
}

func LoadCartridge(path string) *Cartridge {
	f, err := os.Open(path)
	check(err)
	defer f.Close()

	c := Cartridge{}
	c.header = INesHeader{}
	err = binary.Read(f, binary.LittleEndian, &c.header)
	check(err)

	if c.header.MagicNumber != 0x1a53454e {
		panic("Invalid iNES file")
	}

	c.mapperID = int((c.header.Flag7 & 0xF0) | (c.header.Flag6 >> 4))
	c.mirrorMode = int((c.header.Flag6 & 0x1) | (c.header.Flag6 & 0x8 >> 2))

	// read and discard trainer
	if c.header.Flag6&0x4 > 0 {
		_, err = io.ReadFull(f, make([]byte, 512))
		check(err)
	}

	// read PRG rom
	c.prg = make([]byte, int(c.header.SizeRomPRG)*16384)
	_, err = io.ReadFull(f, c.prg)
	check(err)

	// read CHR rom (0 means 8192 bytes of battery backed RAM?)
	sizeCHR := int(c.header.SizeRomCHR)
	if sizeCHR == 0 {
		sizeCHR = 1
	}
	c.chr = make([]byte, sizeCHR*8192)
	_, err = io.ReadFull(f, c.chr)
	check(err)

	return &c
}

func (cartridge *Cartridge) CRC32() (prg, chr, total uint32) {
	prg, chr = crc32.ChecksumIEEE(cartridge.prg), crc32.ChecksumIEEE(cartridge.chr)
	total = crc32.ChecksumIEEE(append(cartridge.prg, cartridge.chr...))
	return
}
