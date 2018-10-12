AlertResponder
=================

**AlertResponder** is a serverless framework for automatic response of security alert.

Overview
------------------

AlertResponder receives an alert that is event of interest from security view point and responses the alert automatically. AlertResponder has 3 parts of automatic response.

1. **Inspector** investigates entities that are appeaered in the alert including IP address, Domain name and store a result: reputation, history of malicious activities, associated cloud instance and etc. Following components are already provided to integrate with your AlertResponder environment. Also you can create own inspector to check logs that is stored into original log storage or log search system.
    - [VirusTotalInspector](https://github.com/m-mizutani/VirusTotalInspector)
2. **Reviewer** receives the alert with result(s) of Inspector and evaluate severity of the alert. Reviewer should be written by each security operator/administrator of your organization because security policies are differ from organazation to organization.
3. **Emitter** finally receives the alert with result of Reviewer's severity evaluation. After that, Emitter sends external integrated system. E.g. PagerDuty, Slack, Github Enterprise, etc. Also automatic quarantine can be configured by AWS Lambda function.
    - [GheReporter](https://github.com/m-mizutani/GheReporter)

![concept](https://user-images.githubusercontent.com/605953/46706573-33aa4380-cc70-11e8-91f8-cc97578f94c4.png)

### Concept

- **Pull based correlation analysis**
- **Alert aggregation**
- **Pluggable Inspectors and Emitters**

Getting Started
------------------

Please replace follwoing variables according to your environment:

- `$REGION`: Replace it with your AWS region. (e.g. ap-northeast-1)
- `$STACK_NAME`: Replace it with CloudFormation stack name

```bash
$ curl -o alert_responder.yml https://s3-$REGION.amazonaws.com/cfn-assets.$REGION/AlertResponder/templates/latest.yml
$ aws cloudformation deploy --template-file alert_responder.yml --stack-name $STACK_NAME --capabilities CAPABILITY_IAM
```

Development
------------------

### Architecture Overview

![architecture](https://user-images.githubusercontent.com/605953/46709133-4b3bf900-cc7d-11e8-8927-b8f068072f58.png)

### Prerequisite

- awscli >= 1.16.20
- Go >= 1.11
- dep >= v0.5.0
- GNU automake >= 1.16.1

### Deploy and Test

Prepare a parameter file, e.g. `param.cfg` and run make command.

```bash
$ cat param.cfg
StackName=alert-responder-test
CodeS3Bucket=some-bucket
CodeS3Prefix=functions

InspectionDelay=1
ReviewDelay=10
$ make CONFIG=param.cfg deploy
```

NOTE: Please make sure that you need AWS credentials (e.g. API key) and appropriate permissions.

After deploying AlertResponder, you need to deploy emitter plugin for test.

```bash
$ cd test/emitter
$ ./deploy.sh alert-responder-test alert-responder-emitter-test some-bucket functions
```

More details of `deploy.sh`'s arguments are following.

```
$ ./deploy.sh <AlertResponderStackName> <TestEmitterStackName> <S3Bucket> <S3Prefix>
```

- `AlertResponderStackName`: AWS CloudFormation stack name of AlertResponder that you want to integrate the test emitter to.
- `TestEmitterStackName`: AWS CloudFormation stack name of test emitter that you will deploy.
- `S3Bucket`: S3 bucket name for storing test code
- `S3Prefix`: Prefix of S3 key for storing test code

`deploy.sh` creates a test paramter file `test/emitter/test.json`. It's required for integration test.

Then, you can run integration test at top level directory of the git repository.

```
$ go test -v
=== RUN   TestInvokeBySns
--- PASS: TestInvokeBySns (3.39s)
           (snip)
PASS
ok      github.com/m-mizutani/AlertResponder    20.110s
```
