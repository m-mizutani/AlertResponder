package main

import (
	"os"

	ar "github.com/m-mizutani/AlertResponder/lib"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func startInspection(task ar.Task) (*ar.ReportPage, error) {
	logger.WithField("task", task).Info("Start inspection")

	page := ar.ReportPage{
		AlliedHosts: []ar.ReportAlliedHost{
			ar.ReportAlliedHost{
				HostName: []string{"test-host-name"},
			},
		},
	}

	logger.WithField("page", page).Info("Done")

	return &page, nil
}

func main() {
	funcName := os.Getenv("SUBMITTER")
	region := os.Getenv("AWS_REGION")

	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	ar.Inspect(startInspection, funcName, region)
}
