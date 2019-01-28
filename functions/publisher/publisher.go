package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/AlertResponder/lib"
)

var logger = logrus.New()

type parameters struct {
	region             string
	reportNotification string
}

func buildParameters(ctx context.Context) (*parameters, error) {
	arn, err := lib.NewArnFromContext(ctx)
	if err != nil {
		return nil, err
	}

	params := parameters{
		region:             arn.Region(),
		reportNotification: os.Getenv("REPORT_NOTIFICATION"),
	}

	return &params, nil
}

// HandleRequest is Lambda handler
func HandleRequest(ctx context.Context, report lib.Report) (string, error) {
	logger.WithField("report", report).Info("Start")

	params, err := buildParameters(ctx)
	if err != nil {
		return "", err
	}

	report.Status = lib.StatusPublished
	err = lib.PublishSnsMessage(params.reportNotification, params.region, report)
	if err != nil {
		return "Error", err
	}

	return "Good", nil
}

func main() {
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	lambda.Start(HandleRequest)
}
