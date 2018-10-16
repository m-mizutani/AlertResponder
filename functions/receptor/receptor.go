package main

import (
	"context"
	"encoding/json"

	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Config is data structure for emitter main procedure
type Config struct {
	Region         string
	TaskStreamName string
	AlertMapName   string
	ReportTo       string
}

type ReceptorResponse struct {
	ReportIDs []string `json:"report_ids"`
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

func ParseSnsEvent(event events.SNSEvent) ([]lib.Alert, error) {
	alerts := []lib.Alert{}

	for _, record := range event.Records {
		src := record.SNS.Message
		log.Println("data = ", src)

		alert := lib.Alert{}
		err := json.Unmarshal([]byte(src), &alert)
		if err != nil {
			log.Println("Invalid alert data: ", string(src))
			return alerts, errors.Wrap(err, "Invalid json format in SNS message")
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
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

func alertToReport(cfg Config, alert lib.Alert) (lib.Report, error) {
	lib.Dump("alert", alert)
	alertMap := NewAlertMap(cfg.AlertMapName, cfg.Region)

	reportID, isNew, err := alertMap.sync(alert)
	if err != nil {
		return lib.Report{}, err
	}
	report := lib.NewReport(reportID, alert)
	if isNew {
		report.Status = lib.StatusNew
	} else {
		report.Status = lib.StatusOngoing
	}

	return report, nil
}

// Handler is main logic of Emitter
func Handler(cfg Config, alerts []lib.Alert) ([]string, error) {
	log.Printf("Start handling %d alert(s)\n", len(alerts))
	resp := []string{}

	for _, alert := range alerts {
		report, err := alertToReport(cfg, alert)
		if err != nil {
			return resp, err
		}

		err = lib.ExecDelayMachine(os.Getenv("DISPATCH_MACHINE"), cfg.Region, report)
		if err != nil {
			return resp, errors.Wrap(err, "Fail to start DispatchMachine")
		}

		if report.IsNew() {
			err = lib.ExecDelayMachine(os.Getenv("REVIEW_MACHINE"), cfg.Region, report)
			if err != nil {
				return resp, errors.Wrap(err, "Fail to start ReviewMachine")
			}
		}

		report.Status = "new"
		err = lib.PublishSnsMessage(os.Getenv("REPORT_LINE"), cfg.Region, report)
		if err != nil {
			return resp, err
		}

		log.Println("put alert to task stream")
		resp = append(resp, string(report.ID))
	}

	return resp, nil
}

// HandleRequest is Lambda handler
func HandleRequest(ctx context.Context, event events.SNSEvent) (ReceptorResponse, error) {
	lib.Dump("Event", event)

	var resp ReceptorResponse

	cfg, err := buildConfig(ctx)
	if err != nil {
		return resp, err
	}

	events, err := ParseSnsEvent(event)
	if err != nil {
		return resp, err
	}

	ids, err := Handler(*cfg, events)
	if err != nil {
		return resp, err
	}

	resp.ReportIDs = ids
	return resp, nil
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
	lambda.Start(HandleRequest)
}
