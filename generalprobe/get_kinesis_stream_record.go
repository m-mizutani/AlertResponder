package generalprobe

import (
	"encoding/json"
	"github.com/pkg/errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/m-mizutani/AlertResponder/lib"
	log "github.com/sirupsen/logrus"
)

type getKinesisStreamRecord struct {
	logicalID string
	baseScene
	callback GetKinesisStreamRecordCallback
}
type GetKinesisStreamRecordCallback func(data []byte) bool

func GetKinesisStreamRecord(logicalID string, callback GetKinesisStreamRecordCallback) *getKinesisStreamRecord {
	scene := getKinesisStreamRecord{
		logicalID: logicalID,
		callback:  callback,
	}
	return &scene
}

func (x *getKinesisStreamRecord) play() error {
	const maxRetry = 20

	streamName := x.lookupPhysicalID(x.logicalID)

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(x.region()),
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
	for i := 0; i < maxRetry; i++ {
		records, err := kinesisService.GetRecords(&kinesis.GetRecordsInput{
			ShardIterator: shardIter,
		})

		if err != nil {
			log.WithField("records", records).Fatal("Fail to get kinesis records")
		}
		shardIter = records.NextShardIterator

		if len(records.Records) > 0 {
			for _, record := range records.Records {
				if x.callback(record.Data) {
					return nil
				}
			}
		}

		time.Sleep(time.Second * 1)
	}

	return errors.New("No kinesis message")
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
		// To be fix: Now only check first shard of list
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
