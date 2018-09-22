package main

import (
	"context"
	"encoding/json"
	"log"
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

	log.Println("Kinesis PutRecord: ", result)

	return nil
}

// HandleRequest is Lambda handler
func HandleRequest(ctx context.Context, report lib.Report) (string, error) {
	arn, err := lib.NewArnFromContext(ctx)
	if err != nil {
		return "", err
	}

	reportData, err := json.Marshal(&report)

	err = kinesisPutRecord(os.Getenv("STREAM_NAME"), arn.Region(), reportData)
	if err != nil {
		return "", err
	}

	err = lib.ExecDelayMachine(os.Getenv("STATE_MACHINE"), arn.Region(), reportData)
	if err != nil {
		return "", err
	}

	return "done", nil
}

func main() {
	lambda.Start(HandleRequest)
}
