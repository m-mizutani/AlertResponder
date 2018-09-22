package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/AlertResponder/lib"
)

type Config struct {
	Region string
	Github struct {
		Endpoint   string
		Repository string
		Token      string
	}
	ReportTableName string
}

func Handler(cfg *Config, report lib.Report) (string, error) {
	return "done", nil
}

func buildConfig(ctx context.Context) (*Config, error) {
	arn, err := lib.NewArnFromContext(ctx)
	if err != nil {
		return nil, err
	}

	secretID := os.Getenv("SECRET_ID")
	log.Printf("Retrieving secret values of %s from %s\n", secretID, arn.Region())
	values, err := lib.GetSecretValues(secretID, arn.Region())
	if err != nil {
		log.Printf("Fail to retrieve secret values: %s\n", err)
		return nil, err
	}
	log.Printf("Retrieved\n")

	cfg := Config{
		Region:          arn.Region(),
		ReportTableName: os.Getenv("REPORT_TABLE"),
	}
	cfg.Github.Endpoint = values.GitHubEndpoint
	cfg.Github.Repository = values.GitHubRepo
	cfg.Github.Token = values.GitHubToken

	return &cfg, nil
}

func HandleRequest(ctx context.Context, report lib.Report) (string, error) {
	lib.DumpJson("report", report)

	cfg, err := buildConfig(ctx)
	if err != nil {
		return "ng", err
	}

	return Handler(cfg, report)
}

func main() {
	lambda.Start(HandleRequest)
}
