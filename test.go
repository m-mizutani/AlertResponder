package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/satori/go.uuid"

	"github.com/k0kubun/pp"
	"github.com/m-mizutani/AlertResponder/lib"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/lambda"
)

type testParams struct {
	StackName           string `json:"StackName"`
	Region              string `json:"Region"`
	ReportResults   string `json:"ReportResultsArn"`
	TaskStream          string
	ReportData          string
	Receptor            string
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

func getTestParams(configPath string) testParams {
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
				pp.Println("data = ", string(record.Data))

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

func invokeReceptor(funcName, region string, event *events.KinesisEvent) string {
	eventData, err := json.Marshal(event)
	if err != nil {
		log.Fatal("Fail to marshal alert", err)
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

	pp.Println("lambda", resp)

	var msg map[string]interface{}
	err = json.Unmarshal(resp.Payload, &msg)
	if err != nil {
		pp.Printf("Payload = %s\n", string(resp.Payload))
	} else {
		pp.Println(msg)
	}

	var receptorResp receptorResponse
	err = json.Unmarshal(resp.Payload, &receptorResp)
	if err != nil {
		log.Fatal("Fail to unmarshal response", err)
	}
	if len(receptorResp.ReportIDs) != 1 {
		log.Fatal("Invalid length of report ID:", receptorResp)
	}

	return receptorResp.ReportIDs[0]
}

func doIntegrationTest(opt *options) {
	log.Println("=== Start integration test ===")

	params := getTestParams(opt.TestConfig)

	pp.Println("params: ", params)
	alertKey := uuid.NewV4().String()

	pp.Println("AlertKey = ", alertKey)

	alert := lib.Alert{
		Key:  alertKey,
		Rule: "test",
		Attrs: []lib.Attribute{
			lib.Attribute{
				Type:    "ipaddr",
				Value:   "10.0.0.1",
				Key:     "source address",
				Context: "remote",
			},
		},
	}
	alertData, err := json.Marshal(alert)
	if err != nil {
		log.Fatal("Fail to marshal alert:", err)
	}

	event := events.KinesisEvent{
		Records: []events.KinesisEventRecord{
			events.KinesisEventRecord{
				Kinesis: events.KinesisRecord{Data: alertData},
			},
		},
	}

	reportID := invokeReceptor(params.Receptor, params.Region, &event)

	task := readKinesisStream(params.TaskStream, params.Region, reportID)

	pp.Println(task)

	var section lib.Section
	section.Title = "Test"
	section.Text = []string{
		"# Test",
		"",
		"This is test message.",
	}
	remote := lib.ReportRemoteHost{
		IPAddr: []string{task.Attr.Value},
	}
	section.RemoteHost = &remote

	reportData := lib.NewReportData(lib.ReportID(reportID))
	reportData.SetSection(section)

	err = reportData.Submit(params.ReportData, params.Region)
	if err != nil {
		log.Fatal("Fail to submit report data", err)
	}

	log.Println("=== Completed Tests ===")
	return
}
