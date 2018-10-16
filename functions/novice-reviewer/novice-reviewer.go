package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	ar "github.com/m-mizutani/AlertResponder/lib"
)

// HandleRequest is Lambda handler
func HandleRequest(ctx context.Context, report ar.Report) (ar.ReportResult, error) {
	ar.Dump("report", report)

	res := ar.ReportResult{Severity: "unclassified"}

	return res, nil
}

func main() {
	lambda.Start(HandleRequest)
}
