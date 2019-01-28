package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/AlertResponder/lib"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func handleRequest(ctx context.Context, event events.SNSEvent) error {
	logger.WithField("event", event).Info("Start")

	for _, record := range event.Records {
		var report lib.Report
		if err := json.Unmarshal([]byte(record.SNS.Message), &report); err != nil {
			logger.WithError(err).Error("Fail to unmarshal report data")
			return err
		}

		logger.WithField("report", report).Info("Reported")
	}
	return nil
}

func main() {
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})
	lambda.Start(handleRequest)
}
