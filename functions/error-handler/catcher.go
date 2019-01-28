package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

type errorInfo struct {
	Error string `json:"Error"`
	Cause string `json:"Cause"`
}

type ErrorEvent struct {
	ErrorInfo errorInfo `json:"Error"`
}

func handleRequest(ctx context.Context, errEvent ErrorEvent) error {
	logger.WithField("Error", errEvent).Info("Got error")
	return nil
}

func main() {
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	lambda.Start(handleRequest)
}
