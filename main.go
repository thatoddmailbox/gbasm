package main

import (
	"flag"
	"log"
	"os"
	"path"
	"sort"
	"strconv"

	"github.com/thatoddmailbox/gbasm/rom"
	"github.com/thatoddmailbox/gbasm/utils"
)

func main() {
	log.Println("gbasm")

	outputFileName := flag.String("output", "out.gb", "The path and name of the output file.")

	flag.Parse()

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
	outputFile, err := os.OpenFile(*outputFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	_, err = outputFile.Write(rom.Current.Output[:])
	if err != nil {
		panic(err)
	}

	log.Println("Constant listing:")

	definitionKeys := make([]string, len(rom.Current.Definitions))
	i := 0
	for name, _ := range rom.Current.Definitions {
		definitionKeys[i] = name
		i += 1
	}

	sort.Strings(definitionKeys)

	for _, name := range definitionKeys {
		value := rom.Current.Definitions[name]
		log.Println(" *", name, value, "0x"+strconv.FormatInt(int64(value), 16))
	}

	log.Println()
	log.Printf("Usage: %d out of %d bytes", rom.Current.UsedByteCount, len(rom.Current.Output))
}
