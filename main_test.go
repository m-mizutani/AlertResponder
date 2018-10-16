package main_test

// This test is an integration test of AlertResponder,
// not unit test of main module.

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/guregu/dynamo"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"

	"github.com/k0kubun/pp"
	"github.com/m-mizutani/AlertResponder/lib"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sns"
)

var (
	Verbose        = os.Getenv("AR_VERBOSE") != ""
	TestConfigPath = os.Getenv("AR_TEST_CONFIG")
)

type testParams struct {
	StackName         string `json:"StackName"`
	Region            string `json:"Region"`
	ReportResults     string `json:"ReportResultsArn"`
	TaskStream        string
	ReportData        string
	Receptor          string
	AlertMap          string
	AlertNotification string
}

func loadTestConfig(fpath string) testParams {
	fd, err := os.Open(fpath)
	if err != nil {
		log.Fatal("Fail to open TestConfig:", fpath, err)
	}
	defer fd.Close()

	params := testParams{}
	fdata, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Fatal("Fail to read TestConfig:", fpath, err)
	}

	err = json.Unmarshal(fdata, &params)
	if err != nil {
		log.Fatal("Fail to unmarshal TestConfig", fpath, err)
	}

	return params
}

func getTestParams() testParams {
	var configPath string
	if TestConfigPath != "" {
		configPath = TestConfigPath
	} else {
		configPath = filepath.Join("test", "emitter", "test.json")
	}

	params := loadTestConfig(configPath)

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(params.Region),
	}))
	client := cloudformation.New(ssn)

	resp, err := client.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(params.StackName),
	})
	if err != nil {
		log.Fatal("Fail to get CloudFormation Stack resources: ", err)
	}

	for _, resource := range resp.StackResources {
		switch *resource.LogicalResourceId {
		case "TaskStream":
			params.TaskStream = *resource.PhysicalResourceId
		case "ReportData":
			params.ReportData = *resource.PhysicalResourceId
		case "Receptor":
			params.Receptor = *resource.PhysicalResourceId
		case "AlertMap":
			params.AlertMap = *resource.PhysicalResourceId
		case "AlertNotification":
			params.AlertNotification = *resource.PhysicalResourceId
		}
	}

	return params
}

func readKinesisStream(streamName, region, reportID string) *lib.Task {
	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	kinesisService := kinesis.New(ssn)
	resp, err := kinesisService.ListShards(&kinesis.ListShardsInput{
		StreamName: aws.String(streamName),
	})
	if err != nil {
		log.Fatal("Fail to shard list", err)
	}

	shardList := []string{}
	for _, s := range resp.Shards {
		shardList = append(shardList, *s.ShardId)
	}

	if len(shardList) != 1 {
		log.Fatal("Invalid shard number: ", len(shardList), ", expected 1")
	}

	now := time.Now()

	iter, err := kinesisService.GetShardIterator(&kinesis.GetShardIteratorInput{
		ShardId:           aws.String(shardList[0]),
		ShardIteratorType: aws.String("AT_TIMESTAMP"),
		StreamName:        aws.String(streamName),
		Timestamp:         &now,
	})
	if err != nil {
		log.Fatal("Fail to get iterator", err)
	}

	shardIter := iter.ShardIterator
	for i := 0; i < 20; i++ {
		records, err := kinesisService.GetRecords(&kinesis.GetRecordsInput{
			ShardIterator: shardIter,
		})

		if err != nil {
			log.Fatal("Fail to get kinesis records", records)
		}
		shardIter = records.NextShardIterator

		if len(records.Records) > 0 {
			for _, record := range records.Records {
				var task lib.Task
				err := json.Unmarshal(record.Data, &task)
				if err != nil {
					log.Fatal("Fail to unmarshal Task")
				}

				if Verbose {
					pp.Println("data = ", string(record.Data))
				}

				if string(task.ReportID) == reportID {
					return &task
				}
			}
		}

		time.Sleep(time.Second * 1)
	}

	log.Fatal("No kinesis message")
	return nil
}

type receptorResponse struct {
	ReportIDs []string `json:"report_ids"`
}

func invokeReceptor(funcName, region string, alert lib.Alert) (string, error) {
	alertData, err := json.Marshal(alert)
	// pp.Println(string(alertData))
	if err != nil {
		return "", errors.Wrap(err, "marshal alert")
	}

	event := events.SNSEvent{
		Records: []events.SNSEventRecord{
			events.SNSEventRecord{
				SNS: events.SNSEntity{
					Message: string(alertData),
				},
			},
		},
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		return "", errors.Wrap(err, "unmarshal event")
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	lambdaService := lambda.New(ssn)

	resp, err := lambdaService.Invoke(&lambda.InvokeInput{
		FunctionName: aws.String(funcName),
		Payload:      eventData,
	})
	if err != nil {
		log.Fatal("Fail to invoke lambda", err)
	}

	if Verbose {
		pp.Println("lambda", resp)
	}

	var msg map[string]interface{}
	err = json.Unmarshal(resp.Payload, &msg)
	if err != nil {
		if Verbose {
			pp.Printf("Payload = %s\n", string(resp.Payload))
		}
	} else {
		if Verbose {
			pp.Println(msg)
		}
	}

	var receptorResp receptorResponse
	err = json.Unmarshal(resp.Payload, &receptorResp)
	if err != nil {
		pp.Println(string(resp.Payload))
		return "", errors.Wrap(err, "unmarshal receptor's response")
	}
	if len(receptorResp.ReportIDs) != 1 {
		pp.Println(receptorResp)
		return "", errors.Wrap(err, "invalid number of report ID set")
	}

	return receptorResp.ReportIDs[0], nil
}

func publishSnsMessage(topicArn, region string, alert lib.Alert) error {
	data, err := json.Marshal(&alert)
	if err != nil {
		return errors.Wrap(err, "Fail to marshal report data")
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	snsService := sns.New(ssn)

	resp, err := snsService.Publish(&sns.PublishInput{
		Message:  aws.String(string(data)),
		TopicArn: aws.String(topicArn),
	})

	if Verbose {
		pp.Println("SNS response", resp)
	}

	if err != nil {
		return errors.Wrap(err, "Fail to publish report")
	}

	return nil
}

type reportResult struct {
	ReportID  lib.ReportID `dynamo:"report_id"`
	Timestamp float64      `dynamo:"timestamp"`
	Report    []byte       `dynamo:"report"`
}

func genAlert() lib.Alert {
	alertKey := uuid.NewV4().String()
	if Verbose {
		pp.Println("AlertKey = ", alertKey)
	}

	alert := lib.Alert{
		Name: "Test Detection",
		Key:  alertKey,
		Rule: "test",
		Attrs: []lib.Attribute{
			lib.Attribute{
				Type:    "ipaddr",
				Value:   "195.22.26.248",
				Key:     "source address",
				Context: []string{"remote"},
			},
		},
	}

	return alert
}
func TestInvokeBySns(t *testing.T) {
	params := getTestParams()
	if Verbose {
		pp.Println("params: ", params)
	}

	alert := genAlert()
	err := publishSnsMessage(params.AlertNotification, params.Region, alert)
	require.NoError(t, err)
	now := time.Now().UTC()

	time.Sleep(3 * time.Second)

	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(params.Region)})

	type AlertRecord struct {
		AlertID   string       `dynamo:"alert_id"`
		AlertKey  string       `dynamo:"alert_key"`
		Rule      string       `dynamo:"rule"`
		ReportID  lib.ReportID `dynamo:"report_id"`
		AlertData []byte       `dynamo:"alert_data"`
		Timestamp time.Time    `dynamo:"timestamp"`
	}
	var results []AlertRecord
	table := db.Table(params.AlertMap)
	err = table.Scan().Filter("'timestamp' >= ?", now).All(&results)

	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
}

func TestIntegration(t *testing.T) {
	params := getTestParams()
	if Verbose {
		pp.Println("params: ", params)
	}

	alert := genAlert()

	reportID, err := invokeReceptor(params.Receptor, params.Region, alert)
	require.NoError(t, err)

	task := readKinesisStream(params.TaskStream, params.Region, reportID)

	if Verbose {
		pp.Println(task)
	}

	var page lib.ReportPage
	page.Title = "Test"
	remote := lib.ReportRemoteHost{
		IPAddr: []string{task.Attr.Value},
	}
	page.RemoteHost = append(page.RemoteHost, remote)

	cmpt := lib.NewReportComponent(lib.ReportID(reportID))
	cmpt.SetPage(page)

	err = cmpt.Submit(params.ReportData, params.Region)
	require.NoError(t, err)

	// Check eventual result(s)
	time.Sleep(time.Second * 15)

	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(params.Region)})
	table := db.Table(params.ReportResults)

	var results []reportResult
	err = table.Get("report_id", string(reportID)).All(&results)
	require.NoError(t, err)

	assert.Equal(t, 2, len(results))

	reports := []lib.Report{}

	sort.Slice(results, func(i, j int) bool { return results[i].Timestamp < results[j].Timestamp })

	for _, result := range results {
		var report lib.Report
		err := json.Unmarshal(result.Report, &report)
		assert.NoError(t, err)
		reports = append(reports, report)
	}

	assert.Equal(t, 0, len(reports[0].Content.RemoteHosts))
	assert.Equal(t, 0, len(reports[0].Content.LocalHosts))
	assert.NotEqual(t, 0, len(reports[1].Content.RemoteHosts))
	assert.Equal(t, 0, len(reports[1].Content.LocalHosts))

	return
}
