package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/k0kubun/pp"
	"github.com/m-mizutani/AlertResponder/lib"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/lambda"
)

type testParams struct {
	StackName           string `json:"StackName"`
	Region              string `json:"Region"`
	ReportReceiverArn   string `json:"ReportReceiverArn"`
	IncidentReceiverArn string `json:"IncidentReceiverArn"`
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

func readKinesisStream(streamName, region string) {
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

	for i := 0; i < 20; i++ {
		records, err := kinesisService.GetRecords(&kinesis.GetRecordsInput{
			ShardIterator: iter.ShardIterator,
		})
		if err != nil {
			log.Fatal("Fail to get kinesis records", records)
		}
		if len(records.Records) > 0 {
			pp.Println(records.Records)
		}

		time.Sleep(time.Second * 1)
	}

	log.Fatal("No kinesis message")
}

func invokeReceptor(funcName, region string, alert *lib.Alert) {
	alertData, err := json.Marshal(alert)
	if err != nil {
		log.Fatal("Fail to marshal alert", err)
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	lambdaService := lambda.New(ssn)

	resp, err := lambdaService.Invoke(&lambda.InvokeInput{
		FunctionName: aws.String(funcName),
		Payload:      alertData,
	})
	if err != nil {
		log.Fatal("Fail to invoke lambda", err)
	}

	pp.Println("lambda", resp)
	pp.Println("payload:", string(resp.Payload))
}

func doIntegrationTest(opt *options) {
	log.Println("=== Start integration test ===")

	params := getTestParams(opt.TestConfig)

	/*
		ssn := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(opt.Region),
		}))
		pp.Println("sns res = ", ssn)
	*/

	// readKinesisStream(params.TaskStream, params.Region)

	pp.Println("params: ", params)

	alert := lib.Alert{}
	invokeReceptor(params.Receptor, params.Region, &alert)

	readKinesisStream(params.TaskStream, params.Region)

	log.Println("=== Exit integration test ===")
	return
}
