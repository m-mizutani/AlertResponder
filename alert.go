package main

import (
	"encoding/json"
	"io/ioutil"

	uuid "github.com/satori/go.uuid"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func sendAlert(region, topicArn string, alert lib.Alert) error {
	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	snsService := sns.New(ssn)

	data, err := json.Marshal(alert)
	if err != nil {
		log.WithField("alert", alert).WithError(err).Error("Fail to marshal")
		return err
	}
	resp, err := snsService.Publish(&sns.PublishInput{
		Message:  aws.String(string(data)),
		TopicArn: aws.String(topicArn),
	})

	log.WithField("SNS response", resp).Info("Send SNS topic")
	if err != nil {
		return errors.Wrap(err, "Fail to publish report")
	}

	return nil
}

func alertCommand(region, stackName, alertFile string, genAlertKey bool) error {
	// Preapare alert data
	data, err := ioutil.ReadFile(alertFile)
	if err != nil {
		return errors.Wrap(err, alertFile)
	}

	var alert lib.Alert
	err = json.Unmarshal(data, &alert)
	if err != nil {
		return errors.Wrap(err, alertFile)
	}

	topicArn, err := lib.GetPhysicalResourceId(region, stackName, "AlertNotification")
	if err != nil {
		return err
	}

	if genAlertKey {
		alert.Key = uuid.NewV4().String()
		log.WithField("NewAlertKey", alert.Key).Info("Generate a new alert key")
	}

	return sendAlert(region, topicArn, alert)
}
