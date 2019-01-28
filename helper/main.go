package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

type parameters struct {
	LambdaRoleArn    string
	AlertNotifyTopic string
	DlqTopicName     string
}

func appendParam(items []string, key string) []string {
	if v := getValue(key); v != "" {
		return append(items, fmt.Sprintf("%s=%s", key, v))
	}

	return items
}

func getValue(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	configFile := os.Getenv("AR_CONFIG")
	if configFile == "" {
		return ""
	}

	fd, err := os.Open(configFile)
	if err != nil {
		logger.Fatal("Can not open AR_CONFIG: ", configFile, err)
	}
	defer fd.Close()

	raw, err := ioutil.ReadAll(fd)
	if err != nil {
		logger.Fatal("Fail to read AR_CONFIG", err)
	}

	var param map[string]string
	err = json.Unmarshal(raw, &param)
	if err != nil {
		logger.Fatal("Fail to unmarshal config json", err)
	}

	if val, ok := param[key]; ok {
		return val
	}

	return ""
}

func makeParameters() {
	parameterNames := []string{
		"LambdaRoleArn",
		"StepFunctionRoleArn",
		"ReviewerLambdaArn",
		"InspectionDelay",
		"ReviewDelay",
	}

	var items []string
	for _, paramName := range parameterNames {
		items = appendParam(items, paramName)
	}

	if len(items) > 0 {
		fmt.Printf("--parameter-overrides %s", strings.Join(items, " "))
	}
}

func makeTestParameters() {
	region := getValue("Region")
	stackName := getValue("StackName")
	if region == "" {
		logger.Fatal("'Region' parameter is required in config or environment variable.")
	}
	if stackName == "" {
		logger.Fatal("'StackName' parameter is required in config or environment variable.")
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	cfn := cloudformation.New(ssn)
	var nextToken *string

	var items []string
	for {
		resp, err := cfn.ListStackResources(&cloudformation.ListStackResourcesInput{
			NextToken: nextToken,
			StackName: aws.String(stackName),
		})

		if err != nil {
			logger.Fatal("Fail to get list of stack resources: ", err)
		}

		for _, resource := range resp.StackResourceSummaries {
			switch *resource.LogicalResourceId {
			case "TaskNotification":
				items = append(items, fmt.Sprintf("TaskNotification=%s",
					*resource.PhysicalResourceId))
			case "ReportNotification":
				items = append(items, fmt.Sprintf("ReportNotification=%s",
					*resource.PhysicalResourceId))
			case "Submitter":
				items = append(items, fmt.Sprintf("Submitter=%s",
					*resource.PhysicalResourceId))
			}
		}

		if resp.NextToken == nil {
			break
		}

		nextToken = resp.NextToken
	}

	fmt.Printf("--parameter-overrides %s", strings.Join(items, " "))
}

func main() {
	logger.SetLevel(logrus.InfoLevel)

	if len(os.Args) < 2 || 3 < len(os.Args) {
		logger.Fatalf("Usage) %s [mkparam|mktest|get <paramName>]", os.Args[0])
	}

	switch os.Args[1] {
	case "mkparam":
		makeParameters()
	case "mktest":
		makeTestParameters()
	case "get":
		fmt.Print(getValue(os.Args[2]))
	}
}
