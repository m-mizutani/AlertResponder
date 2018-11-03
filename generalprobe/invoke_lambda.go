package generalprobe

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	// "github.com/k0kubun/pp"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type invokeLambda struct {
	logicalID string
	input     []byte
	event     interface{}
	callback  InvokeLambdaCallback
	baseScene
}
type InvokeLambdaCallback func(response []byte)

func InvokeLambda(logicalID string, callback InvokeLambdaCallback) *invokeLambda {
	scene := invokeLambda{
		logicalID: logicalID,
		callback:  callback,
	}
	return &scene
}

func toMessage(msg interface{}) string {
	switch v := msg.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			log.WithField("message", msg).Fatal(err)
		}
		return string(raw)
	}
}

func (x *invokeLambda) SnsMessage(input interface{}) *invokeLambda {
	msg := toMessage(input)
	event := events.SNSEvent{
		Records: []events.SNSEventRecord{
			events.SNSEventRecord{
				SNS: events.SNSEntity{
					Message: msg,
				},
			},
		},
	}
	return x.SetEvent(event)
}

func (x *invokeLambda) SetEvent(event interface{}) *invokeLambda {
	x.event = event
	return x
}

func (x *invokeLambda) play() error {
	eventData, err := json.Marshal(x.event)
	if err != nil {
		return errors.Wrap(err, "unmarshal event")
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(x.region()),
	}))
	lambdaService := lambda.New(ssn)

	lambdaArn := x.lookupPhysicalID(x.logicalID)
	resp, err := lambdaService.Invoke(&lambda.InvokeInput{
		FunctionName: aws.String(lambdaArn),
		Payload:      eventData,
	})
	if err != nil {
		log.Fatal("Fail to invoke lambda", err)
	}

	log.WithField("response", resp).Debug("lamba invoked")

	x.callback(resp.Payload)

	/*
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
	*/

	return nil
}
