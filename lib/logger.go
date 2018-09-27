package lib

import (
	"log"
	"strings"

	"github.com/k0kubun/pp"
)

// Dump output logs for AWS Lambda
func Dump(name string, v interface{}) {
	color := pp.ColorScheme{
		Bool:            pp.NoColor,
		Integer:         pp.NoColor,
		Float:           pp.NoColor,
		String:          pp.NoColor,
		StringQuotation: pp.NoColor,
		EscapedChar:     pp.NoColor,
		FieldName:       pp.NoColor,
		PointerAdress:   pp.NoColor,
		Nil:             pp.NoColor,
		Time:            pp.NoColor,
		StructName:      pp.NoColor,
		ObjectLength:    pp.NoColor,
	}
	pp.SetColorScheme(color)

	line := strings.Replace(pp.Sprintln(v), "\n", "", -1)
	log.Printf("%s = %s\n", name, line)
}
