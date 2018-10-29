package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var alertTimeToLive = time.Second * 86400

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
	TTL       time.Time    `dynamo:"ttl"`
}

func GenAlertKey(alertID, rule string) string {
	data := fmt.Sprintf("%s=====%s", alertID, rule)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}

func (x *AlertMap) sync(alert lib.Alert) (lib.ReportID, bool, error) {
	var reportID lib.ReportID
	var isNew bool

	alertID := GenAlertKey(alert.Key, alert.Rule)
	alertData, err := json.Marshal(alert)
	if err != nil {
		return reportID, isNew, errors.Wrap(err, "Fail to unmarshal alert")
	}

	now := time.Now().UTC()
	ttl := now.Add(alertTimeToLive)

	var records []AlertRecord
	err = x.table.Get("alert_id", alertID).Filter("'TTL' > ?", now).All(&records)

	if err != nil {
		return reportID, isNew, errors.Wrap(err, "Fail to get cache")
	}

	var record AlertRecord
	if len(records) == 0 {
		record = AlertRecord{
			AlertKey: alert.Key,
			AlertID:  alertID,
			Rule:     alert.Rule,
			ReportID: lib.NewReportID(),
		}
		isNew = true
		log.WithField("record", record).Info("New alert is created")
	} else {
		log.WithField("records", records).Info("Existing alert is found")
		record = records[0]
		isNew = false
	}

	record.AlertData = alertData
	record.Timestamp = now
	record.TTL = ttl

	lib.Dump("AlertRecord", record)
	err = x.table.Put(&record).Run()
	if err != nil {
		return reportID, isNew, errors.Wrap(err, "Fail to put alert map")
	}

	return record.ReportID, isNew, nil
}
