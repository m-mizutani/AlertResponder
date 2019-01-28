package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func handleRequest(ctx context.Context, page lib.ReportPage) error {
	tableName := os.Getenv("REPORT_DATA")
	region := os.Getenv("AWS_REGION")

	logger.WithFields(logrus.Fields{
		"page":   page,
		"table":  tableName,
		"region": region,
	}).Info("Submitted")

	reportData := lib.NewReportComponent(page.ReportID)
	reportData.SetPage(page)

	if err := reportData.Submit(tableName, region); err != nil {
		return errors.Wrap(err, "Fail to put report data")
	}

	return nil
}

func main() {
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})
	lambda.Start(handleRequest)
}
