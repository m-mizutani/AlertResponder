package lib

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
)

type Inspector func(task Task) (*Section, error)

func handleRequest(ctx context.Context, event events.KinesisEvent, f Inspector, tableArn string) (string, error) {
	// NOTE: I assume that regions of lambda and dynamoDB are same for now
	arn, err := NewArnFromContext(ctx)
	if err != nil {
		return "", errors.Wrap(err, "Error to extract region from context")
	}
	tableRegion := arn.Region()

	if strings.Index(tableArn, "/") < 0 {
		return "", errors.New("Invalid ARN of DynamoDB (missing '/')")
	}
	tableName := strings.Split(tableArn, "/")[1]

	Dump("event.Records", event.Records)
	for _, record := range event.Records {
		task := Task{}
		err := json.Unmarshal(record.Kinesis.Data, &task)
		if err != nil {
			return "", errors.Wrap(err, "Fail to unmarshal kinesis data")
		}

		section, err := f(task)
		Dump("section", section)

		if err != nil {
			return "", errors.Wrap(err, "Fail to generate section")
		}
		// Skip if no report
		if section == nil {
			continue
		}

		reportData := NewReportData(task.ReportID)
		reportData.SetSection(*section)

		if err := reportData.Submit(tableName, tableRegion); err != nil {
			return "", errors.Wrap(err, "Fail to put report data")
		}
	}
	return "ok", nil
}

// Inspect is a wrapper of inspector
func Inspect(f Inspector, tableArn string) {
	lambda.Start(func(ctx context.Context, event events.KinesisEvent) (string, error) {
		return handleRequest(ctx, event, f, tableArn)
	})
}