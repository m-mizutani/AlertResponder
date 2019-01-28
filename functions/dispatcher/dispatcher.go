package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func handleRequest(ctx context.Context, report lib.Report) error {
	region := os.Getenv("AWS_REGION")
	snsTopic := os.Getenv("TASK_NOTIFICATION")
	logger.WithFields(logrus.Fields{
		"report":   report,
		"snsTopic": snsTopic,
		"region":   region,
	}).Info("Start")

	for _, attr := range report.Alert.Attrs {
		task := lib.Task{
			Attr:     attr,
			ReportID: report.ID,
			Alert:    report.Alert,
		}

		logger.WithField("task", task).Info("Dispatch")
		if err := lib.PublishSnsMessage(snsTopic, region, task); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})
	lambda.Start(handleRequest)
}
