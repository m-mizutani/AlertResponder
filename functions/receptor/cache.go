package main

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/pkg/errors"
)

type AlertMap struct {
	table dynamo.Table
}

func NewAlertMap(tableName, region string) *AlertMap {
	alertMap := AlertMap{}

	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(region)})
	alertMap.table = db.Table(tableName)

	return &alertMap
}

type AlertRecord struct {
	AlertID   string       `dynamo:"alert_id"`
	AlertKey  string       `dynamo:"alert_key"`
	Rule      string       `dynamo:"rule"`
	ReportID  lib.ReportID `dynamo:"report_id"`
	AlertData []byte       `dynamo:"alert_data"`
	Timestamp time.Time    `dynamo:"timestamp"`
}

func GenAlertKey(alertID, rule string) string {
	data := fmt.Sprintf("%s=====%s", alertID, rule)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}

func (x *AlertMap) Lookup(alertKey, rule string) (*lib.ReportID, error) {
	alertID := GenAlertKey(alertKey, rule)

	record := AlertRecord{}
	err := x.table.Get("alert_id", alertID).One(&record)
	if err == dynamo.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "Fail to get alert")
	}

	return &record.ReportID, nil
}

func (x *AlertMap) Create(alertKey, rule string, alertData []byte) (*lib.ReportID, error) {
	alertID := GenAlertKey(alertKey, rule)

	record := AlertRecord{
		AlertKey:  alertKey,
		AlertID:   alertID,
		Rule:      rule,
		ReportID:  lib.NewReportID(),
		AlertData: alertData,
		Timestamp: time.Now().UTC(),
	}

	err := x.table.Put(&record).Run()
	if err != nil {
		return nil, errors.Wrap(err, "Fail to put alert map")
	}

	return &record.ReportID, nil
}
