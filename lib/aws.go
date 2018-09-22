package lib

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/sfn"
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

// DecryptKMS decrypts encypted text by KMS
func DecryptKMS(encrypted, region string) (string, error) {
	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	encBin, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", errors.Wrap(err, "Fail to decode encrypted Github token")
	}

	svc := kms.New(ssn)
	input := &kms.DecryptInput{
		CiphertextBlob: encBin,
	}

	result, err := svc.Decrypt(input)
	if err != nil {
		return "", errors.Wrap(err, "Fail to decrypt Github token")
	}

	return string(result.Plaintext), nil
}

type SecretValues struct {
	GitHubToken    string `json:"github_token"`
	GitHubEndpoint string `json:"github_endpoint"`
	GitHubRepo     string `json:"github_repo"`

	GraylogEndpoint string `json:"graylog_endpoint"`
	GraylogToken    string `json:"graylog_token"`

	VirusTotalToken string `json:"virustotal_token"`

	HybridAnalysisToken string `json:"hybridanalysis_token"`
}

func GetSecretValues(secretID, region string) (*SecretValues, error) {
	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	svc := secretsmanager.New(ssn)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	}

	result, err := svc.GetSecretValue(input)

	if err != nil {
		return nil, errors.Wrap(err, "Fail to retrieve secret values")
	}

	values := SecretValues{}
	err = json.Unmarshal([]byte(*result.SecretString), &values)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to parse secret values as JSON")
	}

	return &values, nil
}

func ExecDelayMachine(stateMachineARN string, region string, data []byte) error {
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
