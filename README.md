AlertResponder
=================

**AlertResponder** is a serverless framework for automatic response of security alert.

Concept
------------------

![concept](https://user-images.githubusercontent.com/605953/46706573-33aa4380-cc70-11e8-91f8-cc97578f94c4.png)


Getting Started
------------------

### Deploy main framework

Please replace follwoing variables according to your environment:

- `$REGION`: Replace it with your AWS region. (e.g. ap-northeast-1)
- `$STACK_NAME`: Replace it with CloudFormation stack name

```bash
$ curl -O https://s3-$REGION.amazonaws.com/cfn-assets.$REGION/AlertResponder/templates/latest.yml
$ aws cloudformation deploy --template-file latest.yml --stack-name $STACK_NAME --capabilities CAPABILITY_IAM
```

Development
------------------

### Architecture Overview

![architecture](https://user-images.githubusercontent.com/605953/46709133-4b3bf900-cc7d-11e8-8927-b8f068072f58.png)
