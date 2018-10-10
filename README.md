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
