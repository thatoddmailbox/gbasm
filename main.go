package main

import (
	"log"
	"os"
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

	log.Println(CurrentROM.Definitions)
}