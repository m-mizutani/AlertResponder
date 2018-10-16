package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	log "github.com/sirupsen/logrus"
)

func getStackName(child *cli.Context) (string, error) {
	c := child.Parent()
	stackName := c.String("stack-name")
	if stackName != "" {
		return stackName, nil
	}

	paramFile := c.String("param-file")
	if paramFile == "" {
		return "", errors.New("No available stack name, either --stack-name or --param-file is required")
	}

	fd, err := os.Open(paramFile)
	if err != nil {
		return "", errors.Wrap(err, "Fail to open param file: "+paramFile)
	}
	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return "", errors.Wrap(err, "Fail to read param file")
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "StackName=") {
			arr := strings.Split(line, "=")
			if len(arr) < 2 {
				return "", errors.New(fmt.Sprintf("Invalid format of StackName in the parameter file: %v", line))
			}

			return arr[1], nil
		}
	}

	return "", errors.New("No available StackName in the parameter file")
}

func main() {
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	app := cli.NewApp()
	app.Name = "arcli"
	app.Usage = "AlertResponder Command Line Interface"
	app.Version = "0.0.1"
	app.Action = func(c *cli.Context) error {
		fmt.Println("Hello friend!")
		return nil
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "stack-name, s",
			Usage: "StackName of CloudFormation",
		},
		cli.StringFlag{
			Name:   "param-file, p",
			Usage:  "Parameter file of AlertResponder",
			EnvVar: "AR_CONFIG",
		},
		cli.StringFlag{
			Name:  "region, r",
			Usage: "AWS region",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "alert",
			Usage: "Send an alert to Stack",
			Action: func(c *cli.Context) error {
				stackName, err := getStackName(c)
				if err != nil {
					return err
				}

				return alertCommand(c.Parent().String("region"), stackName,
					c.String("alert"), c.Bool("gen-alert-key"))
			},

			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "alert, a",
					Usage: "Alert data json file",
				},
				cli.BoolFlag{
					Name:  "gen-alert-key, g",
					Usage: "Generate and replace the alert key with new one",
				},
			},
		},
		{
			Name:  "parameters",
			Usage: "N/A",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
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
