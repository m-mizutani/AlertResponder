package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/AlertResponder/lib"
)

type parameters struct {
	region     string
	reportLine string
}

func buildParameters(ctx context.Context) (*parameters, error) {
	arn, err := lib.NewArnFromContext(ctx)
	if err != nil {
		return nil, err
	}

	params := parameters{
		region:     arn.Region(),
		reportLine: os.Getenv("REPORT_LINE"),
	}

	return &params, nil
}

// HandleRequest is Lambda handler
func HandleRequest(ctx context.Context, report lib.Report) (string, error) {
	lib.Dump("report", report)

	params, err := buildParameters(ctx)
	if err != nil {
		return "", err
	}

	report.Status = "published"
	err = lib.PublishSnsMessage(params.reportLine, params.region, &report)
	if err != nil {
		return "Error", err
	}

	return "Good", nil
}

func main() {
	lambda.Start(HandleRequest)
}
