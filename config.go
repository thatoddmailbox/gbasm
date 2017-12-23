package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

type ROMInfo struct {
	Name string
	SupportsDMG bool
}

func ReadConfigFile(basePath string) {
	filePath := path.Join(basePath, "info.toml")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		panic(errors.New("missing info.toml file"))
	}

	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	if _, err := toml.Decode(string(fileContents), &CurrentROM.Info); err != nil {
		panic(err)
	}
}