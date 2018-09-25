package lib

import (
	"log"
	"strings"

	"github.com/k0kubun/pp"
)

// Dump output logs for AWS Lambda
func Dump(name string, v interface{}) {
	data := pp.Sprintln(v)
	line := strings.Replace(data, "\n", "", -1)
	log.Printf("%s = %s\n", name, line)
}
