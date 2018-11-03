package generalprobe

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type publishSnsMessage struct {
	topicName string
	message   []byte
	baseScene
}

func PublishSnsMessage(topicName string, message []byte) *publishSnsMessage {
	scene := publishSnsMessage{
		topicName: topicName,
		message:   message,
	}
	return &scene
}

func (x *publishSnsMessage) play() error {
	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(x.region()),
	}))
	snsService := sns.New(ssn)

	topicArn := x.lookupPhysicalID(x.topicName)
	resp, err := snsService.Publish(&sns.PublishInput{
		Message:  aws.String(string(x.message)),
		TopicArn: aws.String(topicArn),
	})

	log.WithField("result", resp).Debug("sns:Publish result")

	if err != nil {
		return errors.Wrap(err, "Fail to publish report")
	}

	return nil
}
