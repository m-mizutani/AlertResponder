package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/AlertResponder/lib"
)

type errorInfo struct {
	Error string `json:"Error"`
	Cause string `json:"Cause"`
}

type ErrorEvent struct {
	ErrorInfo errorInfo `json:"Error"`
}

func handleRequest(ctx context.Context, errEvent ErrorEvent) (string, error) {
	lib.Dump("Error", errEvent)
	return "done", nil
}

func main() {
	lambda.Start(handleRequest)
}
