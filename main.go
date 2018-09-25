package main

import (
	"log"

	flags "github.com/jessevdk/go-flags"
)

func main() {
	var opts struct {
		TestConfig string `short:"t" long:"testconfig" description:"Integration test config" default:"itest.json"`
		Region     string `short:"r" long:"region" description:"AWS region"`
		StackName  string `short:"s" long:"stack-name" description:"Stack Name"`
		Verbose    bool   `short:"v" long:"verboes"`
	}

	args, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}
	if len(args) == 0 {
		log.Fatal("no command: [test]")
	}

	switch args[0] {
	case "test":
		doIntegrationTest(opts.StackName, opts.Region, opts.Verbose)
	default:
		log.Fatal("not avaialble command, choose from [test]")
	}
}
