package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	ar "github.com/m-mizutani/AlertResponder/lib"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// HandleRequest is Lambda handler
func HandleRequest(ctx context.Context, report ar.Report) (ar.ReportResult, error) {
	logger.WithField("report", report).Info("Start")

	res := ar.ReportResult{Severity: "unclassified", Reason: "NoviceReviewer"}

	return res, nil
}

func main() {
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	lambda.Start(HandleRequest)
}
