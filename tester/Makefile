TEMPLATE_FILE=template.yml
OUTPUT_FILE=sam.yml
FUNCTIONS=build/inspector build/reporter

build/inspector: ./inspector/*.go $(LIBS)
	env GOARCH=amd64 GOOS=linux go build -o build/inspector ./inspector/
build/reporter: ./reporter/*.go $(LIBS)
	env GOARCH=amd64 GOOS=linux go build -o build/reporter ./reporter/

functions: $(FUNCTIONS)

clean:
	rm $(FUNCTIONS)

sam.yml: $(TEMPLATE_FILE) $(FUNCTIONS)
	aws cloudformation package \
		--template-file $(TEMPLATE_FILE) \
		--s3-bucket $(shell ../build/helper get CodeS3Bucket) \
		--s3-prefix $(shell ../build/helper get CodeS3Prefix) \
		--output-template-file $(OUTPUT_FILE)

deploy: $(OUTPUT_FILE)
	aws cloudformation deploy \
		--region $(shell ../build/helper get Region) \
		--template-file $(OUTPUT_FILE) \
		--stack-name $(shell ../build/helper get TestStackName) \
		--capabilities CAPABILITY_IAM $(shell ../build/helper mktest)
	./dump.sh $(shell ../build/helper get TestStackName)