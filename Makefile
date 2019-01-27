TEMPLATE_FILE=template.yml
OUTPUT_FILE=sam.yml
LIBS=lib/*.go
FUNCTIONS=build/receptor build/dispatcher build/compiler build/publisher build/error-handler build/novice-reviewer

all: cli

cli:
	go build -o arcli

build/helper: helper/*.go
	go build -o build/helper ./helper/

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
build/novice-reviewer: ./functions/novice-reviewer/*.go $(LIBS)
	env GOARCH=amd64 GOOS=linux go build -o build/novice-reviewer ./functions/novice-reviewer/

functions: $(FUNCTIONS)

clean:
	rm $(FUNCTIONS)

test:
	go test -v ./lib/

sam.yml: $(TEMPLATE_FILE) $(FUNCTIONS) build/helper
	aws cloudformation package \
		--template-file $(TEMPLATE_FILE) \
		--s3-bucket $(shell ./build/helper get CodeS3Bucket) \
		--s3-prefix $(shell ./build/helper get CodeS3Prefix) \
		--output-template-file $(OUTPUT_FILE)

deploy: $(OUTPUT_FILE) build/helper
	aws cloudformation deploy \
		--template-file $(OUTPUT_FILE) \
		--stack-name $(shell ./build/helper get StackName) \
		--capabilities CAPABILITY_IAM $(shell ./build/helper mkparam)

