package lib

import (
	"encoding/json"
	"log"
)

// Dump output logs for AWS Lambda
func DumpJson(name string, v interface{}) {
	jdata, err := json.Marshal(v)
	if err != nil {
		log.Printf("Error, %s: %s\n", name, err)
	} else {
		log.Printf("%s = %s\n", name, string(jdata))
	}
}
