package main

import (
	"log"
	"os"
	"path"
	"strconv"

	"github.com/thatoddmailbox/gbasm/rom"
	"github.com/thatoddmailbox/gbasm/utils"
)

func main() {
	log.Println("gbasm")

	workingDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	ReadConfigFile(workingDirectory)

	rom.ValidateParameters()
	rom.Initialize()

	// output the actual data
	Assembler_ParseFile(path.Join(workingDirectory, "main.s"), 0x150, 32*utils.KiB)

	rom.Finalize()

	// output the actual file
	outputFileName := path.Join(workingDirectory, "out.gb")
	outputFile, err := os.OpenFile(outputFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	_, err = outputFile.Write(rom.Current.Output[:])
	if err != nil {
		panic(err)
	}

	log.Println("Constant listing:")
	for name, val := range rom.Current.Definitions {
		log.Println("*", name, val, "0x"+strconv.FormatInt(int64(val), 16))
	}
}
