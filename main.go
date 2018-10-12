package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	flags "github.com/jessevdk/go-flags"
)

type options struct {
	TestConfig string `short:"t" long:"testconfig" description:"Integration test config" default:"test/emitter/test.json"`
	StackName  string `short:"s" long:"stack-name"`
	Region     string `short:"r" long:"region"`
	Verbose    bool   `short:"v" long:"verboes"`
}

func main() {
	opts := options{}

	args, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}
	if len(args) == 0 {
		log.Fatal("no command: [test]")
	}

	switch args[0] {
	case "params":
		outputParameters(opts.StackName, opts.Region)
	default:
		log.Fatal("not avaialble command, choose from [test]")
	}
}

func outputParameters(stackName, region string) {
	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	cfn := cloudformation.New(ssn)
	var nextToken *string

	for {
		resp, err := cfn.ListStackResources(&cloudformation.ListStackResourcesInput{
			NextToken: nextToken,
			StackName: aws.String(stackName),
		})

		if err != nil {
			log.Fatal("Fail to get list of stack resources: ", err)
		}

		for _, resource := range resp.StackResourceSummaries {
			switch *resource.LogicalResourceId {
			case "TaskStream":
				log.Println("taskStream = ", *resource.PhysicalResourceId)
			case "ReportData":
				log.Println("reportData =", *resource.PhysicalResourceId)
			}
		}

		if resp.NextToken == nil {
			break
		}

		nextToken = resp.NextToken
	}

	return
}
