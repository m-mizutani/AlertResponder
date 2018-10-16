package main

import (
	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type exportReportInput struct {
	reportID   string
	region     string
	stackName  string
	outputFile string
}

func exportReport(input exportReportInput) error {
	alertMapArn, err := lib.GetPhysicalResourceId(input.region, input.stackName, "AlertMap")
	if err != nil {
		return err
	}
	reportDataArn, err := lib.GetPhysicalResourceId(input.region, input.stackName, "ReportData")
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"AlertMap":   alertMapArn,
		"ReportData": reportDataArn,
	}).Info("Get resource IDs")
	// Not impolemented

	return errors.New("Not implemeneted")
}
