package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/pkg/errors"
)

type CompiledReport struct {
	Report *lib.Report `json:"report"`
}

type parameters struct {
	region    string
	tableName string
}

func buildParameters(ctx context.Context) (*parameters, error) {
	arn, err := lib.NewArnFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to extract region from ARN")
	}

	params := parameters{
		region:    arn.Region(),
		tableName: os.Getenv("REPORT_DATA"),
	}

	return &params, nil
}

// HandleRequest is a main Lambda handler
func HandleRequest(ctx context.Context, report lib.Report) (*lib.Report, error) {
	params, err := buildParameters(ctx)
	if err != nil {
		return nil, err
	}

	lib.Dump("report", report)

	sections, err := lib.FetchReportData(params.tableName, params.region, report.ID)
	if err != nil {
		return nil, err
	}

	report.Sections = sections
	lib.Dump("sections", sections)

	return &report, nil
}

func main() {
	lambda.Start(HandleRequest)
}
