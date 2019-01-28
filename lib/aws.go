package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

func ExecDelayMachine(stateMachineARN string, region string, report Report) error {
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

	Logger.WithField("response", resp).Info("Done startExecution")

	return nil
}

func PublishSnsMessage(topicArn, region string, data interface{}) error {
	msg, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "Fail to marshal report data")
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	snsService := sns.New(ssn)

	resp, err := snsService.Publish(&sns.PublishInput{
		Message:  aws.String(string(msg)),
		TopicArn: aws.String(topicArn),
	})

	Logger.WithField("response", resp).Info("Done SNS Publish")

	if err != nil {
		return errors.Wrap(err, "Fail to publish report")
	}

	return nil
}

func GetSecretValues(secretArn string, values interface{}) error {
	// sample: arn:aws:secretsmanager:ap-northeast-1:1234567890:secret:mytest
	arn := strings.Split(secretArn, ":")
	if len(arn) != 7 {
		return errors.New(fmt.Sprintf("Invalid SecretsManager ARN format: %s", secretArn))
	}
	region := arn[3]

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	mgr := secretsmanager.New(ssn)

	result, err := mgr.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretArn),
	})

	if err != nil {
		return errors.Wrap(err, "Fail to retrieve secret values")
	}

	err = json.Unmarshal([]byte(*result.SecretString), values)
	if err != nil {
		return errors.Wrap(err, "Fail to parse secret values as JSON")
	}

	return nil
}

func GetPhysicalResourceId(region, stackName, logicalId string) (string, error) {
	Logger.WithFields(logrus.Fields{
		"stackName": stackName,
		"region":    region,
	}).Info("Try to get CFn resources")

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	client := cloudformation.New(ssn)

	resp, err := client.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return "", errors.Wrap(err, stackName)
	}

	Logger.WithField("resources", resp.StackResources).Debug("CFn stacks")
	for _, resource := range resp.StackResources {
		if *resource.LogicalResourceId == logicalId {
			Logger.WithField("resource", resource).Info("Found target resource")
			return *resource.PhysicalResourceId, nil
		}
	}

	return "", errors.New("Target resource is not found in " + stackName)
}
