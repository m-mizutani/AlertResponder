package lib

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	lambdaService "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Inspector is callback function type.
type Inspector func(task Task) (*ReportPage, error)

func handleRequest(ctx context.Context, event events.SNSEvent, f Inspector, funcName, region string) error {
	Logger.WithField("event.Records", event.Records).Info("Start events")
	for _, record := range event.Records {
		task := Task{}
		err := json.Unmarshal([]byte(record.SNS.Message), &task)

		if err != nil {
			return errors.Wrap(err, "Fail to unmarshal kinesis data")
		}

		page, err := f(task)
		Logger.WithField("page", page).Info("Got page")

		if err != nil {
			return errors.Wrap(err, "Fail to generate section")
		}
		// Skip if no report
		if page == nil {
			continue
		}

		// Create payload body
		page.ReportID = task.ReportID
		payload, err := json.Marshal(page)
		if err != nil {
			return errors.Wrap(err, "Fail to marshal Page data")
		}

		ssn := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(region),
		}))
		svc := lambdaService.New(ssn)

		input := &lambdaService.InvokeInput{
			FunctionName: &funcName,
			Payload:      payload,
		}
		resp, err := svc.Invoke(input)

		Logger.WithFields(logrus.Fields{
			"input":    input,
			"response": resp,
			"error":    err,
		}).Info("Invoke Lambda")

		if err != nil {
			return errors.Wrap(err, "Fail to invoke submitter")
		}

	}
	return nil
}

// Inspect is a wrapper of inspector
func Inspect(f Inspector, funcName, region string) {
	lambda.Start(func(ctx context.Context, event events.SNSEvent) error {
		return handleRequest(ctx, event, f, funcName, region)
	})
}

func InspectTest(f Inspector, task Task) (*ReportPage, error) {
	page, err := f(task)
	return page, err
}
