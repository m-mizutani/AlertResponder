package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/AlertResponder/lib"
)

// HandleRequest is Lambda handler
func HandleRequest(ctx context.Context, alert lib.Alert) (string, error) {
	lib.Dump("alert", alert)

	return "Good", nil
}

func main() {
	lambda.Start(HandleRequest)
}
