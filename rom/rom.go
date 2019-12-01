package rom

import (
	"errors"

	"github.com/thatoddmailbox/gbasm/utils"
)

type Info struct {
	Name        string
	SupportsDMG bool
}

type ROM struct {
	Info                 Info
	Output               [32 * utils.KiB]byte
	Definitions          map[string]int
	UnpointedDefinitions []string
}

var LogoBitmap = []byte{0xCE, 0xED, 0x66, 0x66, 0xCC, 0x0D, 0x00, 0x0B, 0x03, 0x73, 0x00, 0x83, 0x00, 0x0C, 0x00, 0x0D, 0x00, 0x08, 0x11, 0x1F, 0x88, 0x89, 0x00, 0x0E, 0xDC, 0xCC, 0x6E, 0xE6, 0xDD, 0xDD, 0xD9, 0x99, 0xBB, 0xBB, 0x67, 0x63, 0x6E, 0x0E, 0xEC, 0xCC, 0xDD, 0xDC, 0x99, 0x9F, 0xBB, 0xB9, 0x33, 0x3E}

var Current ROM

func calculateHeaderChecksum(array []byte) byte {
	result := byte(0)
	for _, n := range array {
		result = result - n - 1
	}
	return result
}

func calculateGlobalChecksum(array []byte) int {
	result := 0
	for _, n := range array {
		result = result + int(n)
	}
	return result
}

// ValidateParameters ensures that the provided ROM info is valid.
func ValidateParameters() {
	if len(Current.Info.Name) > 15 {
		panic(errors.New("Specified name for ROM is too long!"))
	}
}

// Initialize sets up the ROM data with the provided information.
func Initialize() {
	Current.Definitions = map[string]int{}

	// create the header

	// entry point, hardcoded jump to 0x150 for now
	Current.Output[0x100] = 0x00 // NOP
	Current.Output[0x101] = 0xC3 // JP
	Current.Output[0x102] = 0x50 // 0x50
	Current.Output[0x103] = 0x01 // 0x01

	// logo bitmap
	copy(Current.Output[0x104:], LogoBitmap)

	// name
	nameArray := []byte(Current.Info.Name)
	copy(Current.Output[0x134:], nameArray) // left over bytes will be null

	// CGB bit
	if Current.Info.SupportsDMG {
		Current.Output[0x143] = 0x80 // DMG/CGB compatible
	} else {
		Current.Output[0x143] = 0xc0 // CGB exclusive
	}

	// licensee code
	Current.Output[0x144] = 0x30 // "0"
	Current.Output[0x145] = 0x31 // "1"

	Current.Output[0x146] = 0x00 // SGB flag
	Current.Output[0x147] = 0x00 // Cartridge type (set to ROM ONLY for now)
	Current.Output[0x148] = 0x00 // ROM size (set to 32 KiB for now)
	Current.Output[0x149] = 0x00 // RAM size (set to no built-in RAM for now)

	Current.Output[0x14A] = 0x01 // Destination code (set to not-Japan)

	Current.Output[0x14B] = 0x33 // Old licensee code (unused)

	Current.Output[0x14C] = 0x00 // Mask ROM version

	// header checksum (global checksum is done at end)
	Current.Output[0x14D] = calculateHeaderChecksum(Current.Output[0x134:0x14D])
}

// Finalize applies final preparations to the ROM file.
func Finalize() {
	// calculate global checksum
	globalChecksum := calculateGlobalChecksum(Current.Output[:])
	Current.Output[0x14E] = byte((globalChecksum & 0xFF00) >> 8) // upper bits
	Current.Output[0x14F] = byte(globalChecksum & 0xFF)          // lower bits
}
