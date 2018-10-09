AR_CONFIG ?= ${CONFIG}

CODE_S3_BUCKET := $(shell cat $(AR_CONFIG) | grep CodeS3Bucket | cut -d = -f 2)
CODE_S3_PREFIX := $(shell cat $(AR_CONFIG) | grep CodeS3Prefix | cut -d = -f 2)
STACK_NAME := $(shell cat $(AR_CONFIG) | grep StackName | cut -d = -f 2)
PARAMETERS := $(shell cat $(AR_CONFIG) | grep -e LambdaRoleArn -e StepFunctionRoleArn -e PolicyLambdaArn -e InspectionDelay -e ReviewDelay | tr '\n' ' ')
TEMPLATE_FILE=template.yml
LIBS=lib/*.go
FUNCTIONS=build/receptor build/dispatcher build/compiler build/publisher build/error-handler

all: cli

cli:
	go build -o arcli

build/receptor: ./functions/receptor/*.go $(LIBS)
	env GOARCH=amd64 GOOS=linux go build -o build/receptor ./functions/receptor/
build/dispatcher: ./functions/dispatcher/*.go $(LIBS)
	env GOARCH=amd64 GOOS=linux go build -o build/dispatcher ./functions/dispatcher/
build/compiler: ./functions/compiler/*.go $(LIBS)
	env GOARCH=amd64 GOOS=linux go build -o build/compiler ./functions/compiler/
build/publisher: ./functions/publisher/*.go $(LIBS)
	env GOARCH=amd64 GOOS=linux go build -o build/publisher ./functions/publisher/
build/error-handler: ./functions/compiler/*.go $(LIBS)
	env GOARCH=amd64 GOOS=linux go build -o build/error-handler ./functions/error-handler/

functions: $(FUNCTIONS)

clean:
	rm $(FUNCTIONS)

test:
	go test -v ./lib/

sam.yml: $(TEMPLATE_FILE) $(FUNCTIONS)
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
