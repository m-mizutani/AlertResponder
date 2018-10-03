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

// TestConfig is for test preference.
type TestConfig struct {
	GithubEndpoint   string `json:"github_endpoint"`
	GithubRepository string `json:"github_repo"`
	GithubToken      string `json:"github_token"`
	AwsRegion        string `json:"aws_region"`
	SecretID         string `json:"secret_id"`
}

// LoadTestConfig provides config data from "test.json".
// The method searches "test.json" toward upper directory
func LoadTestConfig() TestConfig {
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

	cfg := TestConfig{}
	err = json.Unmarshal(rawData, &cfg)
	return cfg
}
