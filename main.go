package main

import (
	"log"

	flags "github.com/jessevdk/go-flags"
)

type options struct {
	TestConfig string `short:"t" long:"testconfig" description:"Integration test config" default:"test/emitter/test.json"`
	Verbose    bool   `short:"v" long:"verboes"`
}

func main() {
	opts := options{}

	args, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}
	if len(args) == 0 {
		log.Fatal("no command: [test]")
	}

	switch args[0] {
	case "test":
		doIntegrationTest(&opts)
	default:
		log.Fatal("not avaialble command, choose from [test]")
	}
}
