package lib

import (
	"log"
	"strings"

	"github.com/k0kubun/pp"
)

// Dump output logs for AWS Lambda
func Dump(name string, v interface{}) {
	coloring := pp.ColoringEnabled
	pp.ColoringEnabled = false
	defer func() { pp.ColoringEnabled = coloring }()

	pp.WithLineInfo = true

	line := strings.Replace(pp.Sprintln(v), "\n", "", -1)
	log.Printf("%s = %s\n", name, line)
}
