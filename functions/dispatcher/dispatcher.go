package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/m-mizutani/AlertResponder/lib"
)

func kinesisPutRecord(streamName, region string, alertData []byte) error {
	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	kinesisClient := kinesis.New(ssn)

	partitionKey := "fixed_value_because_number_of_shard_is_1"
	kinesisInput := kinesis.PutRecordInput{
		Data:         alertData,
		PartitionKey: &partitionKey,
		StreamName:   &streamName,
	}
	result, err := kinesisClient.PutRecord(&kinesisInput)

	if err != nil {
		return err
	}

	lib.Dump("Kinesis PutRecord", result)

	return nil
}

// HandleRequest is Lambda handler
func HandleRequest(ctx context.Context, report lib.Report) (string, error) {
	arn, err := lib.NewArnFromContext(ctx)
	if err != nil {
		return "", err
	}

	lib.Dump("report", report)

	for _, attr := range report.Alert.Attrs {
		task := lib.Task{
			Attr:     attr,
			ReportID: report.ID,
		}

		lib.Dump("task", task)
		taskData, err := json.Marshal(&task)
		err = kinesisPutRecord(os.Getenv("STREAM_NAME"), arn.Region(), taskData)
		if err != nil {
			return "", err
		}
	}

	return "done", nil
}

func main() {
	lambda.Start(HandleRequest)
}
