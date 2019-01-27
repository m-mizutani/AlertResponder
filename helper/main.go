package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type parameters struct {
	LambdaRoleArn    string
	AlertNotifyTopic string
	DlqTopicName     string
}

func appendParam(items []string, key string) []string {
	if v := getValue(key); v != "" {
		return append(items, fmt.Sprintf("%s=%s", key, v))
	}

	return items
}

func getValue(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	configFile := os.Getenv("AR_CONFIG")
	if configFile == "" {
		return ""
	}

	fd, err := os.Open(configFile)
	if err != nil {
		log.Fatal("Can not open CONFIG_FILE: ", configFile, err)
	}
	defer fd.Close()

	raw, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Fatal("Fail to read CONFIG_FILE", err)
	}

	var param map[string]string
	err = json.Unmarshal(raw, &param)
	if err != nil {
		log.Fatal("Fail to unmarshal config json", err)
	}

	if val, ok := param[key]; ok {
		return val
	}

	return ""
}

func makeParameters() {
	parameterNames := []string{
		"LambdaRoleArn",
		"StepFunctionRoleArn",
		"ReviewerLambdaArn",
		"InspectionDelay",
		"ReviewDelay",
	}

	var items []string
	for _, paramName := range parameterNames {
		items = appendParam(items, paramName)
	}

	if len(items) > 0 {
		fmt.Printf("--parameter-overrides %s", strings.Join(items, " "))
	}
}

func main() {
	if len(os.Args) < 2 || 3 < len(os.Args) {
		log.Fatalf("Usage) %s [mkparam|get <paramName>]", os.Args[0])
	}

	switch os.Args[1] {
	case "mkparam":
		makeParameters()
	case "get":
		fmt.Print(getValue(os.Args[2]))
	}
}
