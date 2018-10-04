package lib

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Test helper

// LoadTestConfig provides config data from "test.json".
// The method searches "test.json" toward upper directory
func LoadTestConfig(cfg interface{}) {
	cwd := os.Getenv("PWD")
	var fp *os.File
	var err error

	for cwd != "/" {
		cfgPath := filepath.Join(cwd, "test.json")

		cwd, _ = filepath.Split(strings.TrimRight(cwd, string(filepath.Separator)))

		fp, err = os.Open(cfgPath)
		if err == nil {
			break
		}
	}

	if fp == nil {
		log.Fatal("test.json is not found")
	}

	rawData, err := ioutil.ReadAll(fp)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(rawData, cfg)
	if err != nil {
		panic(err)
	}

	return
}
