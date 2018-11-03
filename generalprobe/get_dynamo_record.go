package generalprobe

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/pkg/errors"
)

type TargetType int

const (
	LogicalID TargetType = iota
	ARN
)

type getDynamoRecord struct {
	target     string
	targetType TargetType

	callback GetDynamoRecordCallback
	baseScene
}
type GetDynamoRecordCallback func(table dynamo.Table) (bool, error)

func GetDynamoRecord(target string, callback GetDynamoRecordCallback) *getDynamoRecord {
	scene := getDynamoRecord{
		target:     target,
		targetType: LogicalID,
		callback:   callback,
	}
	return &scene
}

func (x *getDynamoRecord) SetTargetType(t TargetType) *getDynamoRecord {
	x.targetType = t
	return x
}

func (x *getDynamoRecord) play() error {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(x.region())})
	tableName := x.target
	if x.targetType == LogicalID {
		tableName = x.lookupPhysicalID(x.target)
	}
	table := db.Table(tableName)
	const maxRetry int = 30

	for n := 0; n < maxRetry; n++ {
		time.Sleep(time.Second * 2)

		fetched, err := x.callback(table)
		if err != nil {
			return err
		}
		if fetched {
			return nil
		}
	}

	return errors.New("Timeout to fetch records from DynamoDB")
}
