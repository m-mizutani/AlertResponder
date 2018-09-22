package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	lambdaService "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/pkg/errors"
)

// Config is data structure for emitter main procedure
type Config struct {
	Region         string
	TaskStreamName string
	AlertMapName   string
	ReportTo       string
}

func buildConfig(ctx context.Context) (*Config, error) {
	arn, err := lib.NewArnFromContext(ctx)
	if err != nil {
		return nil, err
	}

	cfg := Config{
		Region:         arn.Region(),
		AlertMapName:   os.Getenv("ALERT_MAP"),
		TaskStreamName: os.Getenv("STREAM_NAME"),
		ReportTo:       os.Getenv("REPORT_TO"),
	}

	return &cfg, nil
}

func ParseEvent(event events.KinesisEvent) ([]lib.Alert, error) {
	alerts := []lib.Alert{}

	for _, record := range event.Records {
		src := record.Kinesis.Data
		log.Println("data = ", string(src))

		alert := lib.Alert{}
		err := json.Unmarshal(src, &alert)
		if err != nil {
			log.Println("Invalid alert data: ", string(src))
			return alerts, errors.Wrap(err, "Invalid json format in KinesisRecord")
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// Handler is main logic of Emitter
func Handler(cfg Config, alerts []lib.Alert) (string, error) {
	log.Printf("Start handling %d alert(s)\n", len(alerts))

	client := lambdaService.New(session.New(), &aws.Config{Region: aws.String(cfg.Region)})

	alertMap := NewAlertMap(cfg.AlertMapName, cfg.Region)

	for _, alert := range alerts {
		log.Printf("alert = %v\n", alert.Alert)

		reportID, err := alertMap.Lookup(alert.Alert.ID, alert.Alert.Rule)
		if err != nil {
			return "", err
		}

		if reportID == nil {
			// Existing alert issue is not found
			reportID, err = alertMap.Create(alert.Alert.ID, alert.Alert.Rule)

			if err != nil {
				return "", errors.Wrap(err, "Failt to create a new alert map")
			}
			log.Printf("Created a new reportDI: %s", *reportID)
		}

		report := lib.NewReport(*reportID, &alert)

		reportData, err := json.Marshal(&report)
		if err != nil {
			return "", errors.Wrap(err, "Fail to marshal report data")
		}

		result, err := client.InvokeAsync(&lambdaService.InvokeAsyncInput{
			FunctionName: aws.String(cfg.ReportTo),
			InvokeArgs:   aws.ReadSeekCloser(bytes.NewReader(reportData)),
		})
		log.Println("invoke result: ", result)

		if err != nil {
			return "", errors.Wrap(err, "Fail to invoke report-to lambda")
		}

		err = lib.ExecDelayMachine(os.Getenv("STATE_MACHINE"), cfg.Region, reportData)
		if err != nil {
			return "", errors.Wrap(err, "Fail to start StateMachine")
		}

		log.Println("put alert to task stream")
	}

	return "ok", nil
}

// HandleRequest is Lambda handler
func HandleRequest(ctx context.Context, event events.KinesisEvent) (string, error) {
	log.Println("Event = ", event)

	cfg, err := buildConfig(ctx)
	if err != nil {
		return "ng", err
	}

	events, err := ParseEvent(event)
	if err != nil {
		return "ng", err
	}

	return Handler(*cfg, events)
}

func main() {
	lambda.Start(HandleRequest)
}
