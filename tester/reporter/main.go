package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func handleRequest(ctx context.Context, report lib.Report) error {
	logger.WithField("report", report).Info("Reported")
	return nil
}

func main() {
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})
	lambda.Start(handleRequest)
}
