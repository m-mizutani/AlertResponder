package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/AlertResponder/lib"
)

// HandleRequest is Lambda handler
func HandleRequest(ctx context.Context, alert lib.Alert) (string, error) {
	log.Println(alert)

	return "Yes, Yes, Yes. Oh my god", nil
}

func main() {
	lambda.Start(HandleRequest)
}
