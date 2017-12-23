package main

import (
	"errors"
	"os"
	"path"
)

type ROM struct {
	Info ROMInfo
	Output [32*KiB]byte
	Definitions map[string]int
	UnpointedDefinitions []string
}

var LogoBitmap = []byte{0xCE, 0xED, 0x66, 0x66, 0xCC, 0x0D, 0x00, 0x0B, 0x03, 0x73, 0x00, 0x83, 0x00, 0x0C, 0x00, 0x0D, 0x00, 0x08, 0x11, 0x1F, 0x88, 0x89, 0x00, 0x0E, 0xDC, 0xCC, 0x6E, 0xE6, 0xDD, 0xDD, 0xD9, 0x99, 0xBB, 0xBB, 0x67, 0x63, 0x6E, 0x0E, 0xEC, 0xCC, 0xDD, 0xDC, 0x99, 0x9F, 0xBB, 0xB9, 0x33, 0x3E}

var CurrentROM ROM

func ROM_Create(basePath string) {
	CurrentROM.Definitions = map[string]int{}

	outputFileName := path.Join(basePath, "out.gb")

	// create the header

	// entry point, hardcoded jump to 0x150 for now
	CurrentROM.Output[0x100] = 0x00 // NOP
	CurrentROM.Output[0x101] = 0xC3 // JP
	CurrentROM.Output[0x102] = 0x50 // 0x50
	CurrentROM.Output[0x103] = 0x01 // 0x01

	// logo bitmap
	copy(CurrentROM.Output[0x104:], LogoBitmap)

	// name
	nameArray := []byte(CurrentROM.Info.Name)
	copy(CurrentROM.Output[0x134:], nameArray) // left over bytes will be null

	// CGB bit
	if CurrentROM.Info.SupportsDMG {
		CurrentROM.Output[0x143] = 0x80 // DMG/CGB compatible
	} else {
		CurrentROM.Output[0x143] = 0xc0 // CGB exclusive
	}

	// licensee code
	CurrentROM.Output[0x144] = 0x30 // "0"
	CurrentROM.Output[0x145] = 0x31 // "1"

	CurrentROM.Output[0x146] = 0x00 // SGB flag
	CurrentROM.Output[0x147] = 0x00 // Cartridge type (set to ROM ONLY for now)
	CurrentROM.Output[0x148] = 0x00 // ROM size (set to 32 KiB for now)
	CurrentROM.Output[0x149] = 0x00 // RAM size (set to no built-in RAM for now)

	CurrentROM.Output[0x14A] = 0x01 // Destination code (set to not-Japan)

	CurrentROM.Output[0x14B] = 0x33 // Old licensee code (unused)

	CurrentROM.Output[0x14C] = 0x00 // Mask ROM version

	// header checksum (global checksum is done at end)
	CurrentROM.Output[0x14D] = ROM_CalculateHeaderChecksum(CurrentROM.Output[0x134:0x14D])

	// output the actual data
	Assembler_ParseFile(path.Join(basePath, "main.s"), 0x150, 32*KiB)

	// calculate global checksum
	globalChecksum := ROM_CalculateGlobalChecksum(CurrentROM.Output[:])
	CurrentROM.Output[0x14E] = byte((globalChecksum & 0xFF00) >> 8) // upper bits
	CurrentROM.Output[0x14F] = byte(globalChecksum & 0xFF) // lower bits

	// output the actual file
	outputFile, err := os.OpenFile(outputFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil { panic(err) }
	defer outputFile.Close()
	_, err = outputFile.Write(CurrentROM.Output[:])
	if err != nil { panic(err) }
}

func ROM_ValidateParameters() {
	if len(CurrentROM.Info.Name) > 15 {
		panic(errors.New("Specified name for ROM is too long!"))
	}
}

func ROM_CalculateHeaderChecksum(array []byte) (byte) {
	result := byte(0)
	for _, n := range array {
		result = result - n - 1
	}
	return result
}

func ROM_CalculateGlobalChecksum(array []byte) (int) {
	result := 0
	for _, n := range array {
		result = result + int(n)
	}
	return result
}