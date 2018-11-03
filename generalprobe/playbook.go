package generalprobe

import (
	// "encoding/json"
	// "time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	// "github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	// "github.com/m-mizutani/AlertResponder/lib"
	log "github.com/sirupsen/logrus"
)

type Playbook struct {
	awsRegion string
	stackName string
	scenes    []Scene
	resources []*cloudformation.StackResource
}

func NewPlaybook(awsRegion, stackName string) Playbook {
	book := Playbook{
		awsRegion: awsRegion,
		stackName: stackName,
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	}))
	client := cloudformation.New(ssn)

	resp, err := client.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		log.Fatal("Fail to get CloudFormation Stack resources: ", err)
	}

	book.resources = resp.StackResources

	return book
}

func (x *Playbook) lookup(logicalId string) string {
	for _, resource := range x.resources {
		if resource.LogicalResourceId != nil && *resource.LogicalResourceId == logicalId {
			return *resource.PhysicalResourceId
		}
	}
	return ""
}

func (x *Playbook) SetScenario(newScenes []Scene) {
	for _, scene := range newScenes {
		scene.setPlaybook(x)
		x.scenes = append(x.scenes, scene)
	}
}

func (x *Playbook) Act() error {
	for _, scene := range x.scenes {
		if err := scene.play(); err != nil {
			return err
		}
	}

	return nil
}

type Scene interface {
	play() error
	setPlaybook(book *Playbook)
}

type baseScene struct {
	playbook *Playbook
}

func (x *baseScene) setPlaybook(book *Playbook)               { x.playbook = book }
func (x *baseScene) region() string                           { return x.playbook.awsRegion }
func (x *baseScene) lookupPhysicalID(logicalID string) string { return x.playbook.lookup(logicalID) }
