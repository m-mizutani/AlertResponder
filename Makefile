AR_CONFIG ?= "param.cfg"

CODE_S3_BUCKET := $(shell cat $(AR_CONFIG) | grep CodeS3Bucket | cut -d = -f 2)
CODE_S3_PREFIX := $(shell cat $(AR_CONFIG) | grep CodeS3Prefix | cut -d = -f 2)
STACK_NAME := $(shell cat $(AR_CONFIG) | grep StackName | cut -d = -f 2)
PARAMETERS := $(shell cat $(AR_CONFIG) | grep -e LambdaRoleArn -e StepFunctionRoleArn -e LambdaArn -e SecretId -e VpcSecurityGroups -e VpcSubnetIds -e NotifyStreamArn | tr '\n' ' ')
TEMPLATE_FILE=template.yml
LIBS=lib/*.go

build/receptor: ./functions/receptor/*.go $(LIBS)
	env GOARCH=amd64 GOOS=linux go build -o build/receptor ./functions/receptor/

build/dispatcher: ./functions/dispatcher/*.go $(LIBS)
	env GOARCH=amd64 GOOS=linux go build -o build/dispatcher ./functions/dispatcher/

build/reviewer: ./functions/reviewer/*.go $(LIBS)
	env GOARCH=amd64 GOOS=linux go build -o build/reviewer ./functions/reviewer/

build/ghe-emitter: ./emitters/ghe/*.go $(LIBS)
	env GOARCH=amd64 GOOS=linux go build -o build/ghe-emitter ./emitters/ghe/

test:
	go test -v ./lib/

sam.yml: build/receptor build/dispatcher build/reviewer build/ghe-emitter template.yml
	aws cloudformation package \
		--template-file $(TEMPLATE_FILE) \
		--s3-bucket $(CODE_S3_BUCKET) \
		--s3-prefix $(CODE_S3_PREFIX) \
		--output-template-file sam.yml

deploy: sam.yml
	aws cloudformation deploy \
		--template-file sam.yml \
		--stack-name $(STACK_NAME) \
		--capabilities CAPABILITY_IAM \
		--parameter-overrides $(PARAMETERS)
