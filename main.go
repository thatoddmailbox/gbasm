package main

import (
	"log"
	"os"
	"strconv"
)

const KiB = 1024

func main() {
	log.Println("gbasm")

	workingDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	ReadConfigFile(workingDirectory)

	ROM_ValidateParameters()

	ROM_Create(workingDirectory)

	log.Println("Constant listing:")
	for name, val := range CurrentROM.Definitions {
		log.Println("*", name, val, "0x" + strconv.FormatInt(int64(val), 16))
	}
}