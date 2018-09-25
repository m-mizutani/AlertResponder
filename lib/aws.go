package lib

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/pkg/errors"
)

// Event is a schema between Pusher and Detector
type Event struct {
	Bucket    string    `json:"s3_bucket"`
	Key       string    `json:"s3_key"`
	EventTime time.Time `json:"time"`
}

// Arn is an utility for AWS Resource Namespace
type Arn struct {
	arn  string
	args []string
}

// Region returns AWS region
func (x Arn) Region() string {
	return x.args[3]
}

// FuncName returns AWS function name of lambda
func (x Arn) FuncName() string {
	return x.args[6]
}

// NewArnFromContext creates Arn instance with context.Context for Lambda function
func NewArnFromContext(ctx context.Context) (Arn, error) {
	lambdaCtx, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return Arn{}, errors.New("Invalid context")
	}

	return NewArn(lambdaCtx.InvokedFunctionArn), nil
}

// NewArn is a constructor of Arn
func NewArn(arn string) Arn {
	obj := Arn{arn: arn}
	obj.args = strings.Split(arn, ":")
	return obj
}

func ExecDelayMachine(stateMachineARN string, region string, report *Report) error {
	data, err := json.Marshal(report)
	if err != nil {
		return errors.Wrap(err, "Fail to marshal report data")
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	svc := sfn.New(ssn)

	input := sfn.StartExecutionInput{
		Input:           aws.String(string(data)),
		StateMachineArn: aws.String(stateMachineARN),
	}
	resp, err := svc.StartExecution(&input)
	if err != nil {
		return err
	}

	log.Println(resp)

	return nil
}

func PublishSnsMessage(topicArn, region string, report *Report) error {
	data, err := json.Marshal(report)
	if err != nil {
		return errors.Wrap(err, "Fail to marshal report data")
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	snsService := sns.New(ssn)

	resp, err := snsService.Publish(&sns.PublishInput{
		Message:  aws.String(string(data)),
		TopicArn: aws.String(topicArn),
	})

	Dump("SNS response", resp)

	if err != nil {
		return errors.Wrap(err, "Fail to publish report")
	}

	return nil
}
