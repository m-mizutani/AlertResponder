package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/AlertResponder/lib"
)

// HandleRequest is Lambda handler
func HandleRequest(ctx context.Context, v interface{}) (string, error) {
	lib.Dump("data", v)

	return "Good", nil
}

func main() {
	lambda.Start(HandleRequest)
}
