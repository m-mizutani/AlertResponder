#!/bin/bash

TEMPLATE_FILE=template.yml
OUTPUT_FILE=`mktemp`

if [ $# -le 0 ]; then
    echo "usage) $0 StackName CodeS3Bucket CodeS3Prefix [LambdaArn]"
    exit 1
fi

if [ "$5" != "" ]; then
   OPT="--parameter-overrides LambdaArn=$2"
else
   OPT=""
fi

echo "Output Template: $OUTPUT_FILE"
echo ""

aws cloudformation package --template-file $TEMPLATE_FILE --output-template-file $OUTPUT_FILE --s3-bucket $2 --s3-prefix $3
aws cloudformation deploy --template-file $OUTPUT_FILE --stack-name $1 --capabilities CAPABILITY_IAM $OPT

echo ""
echo ""

Resource=`aws cloudformation describe-stack-resources --stack-name $1 | jq '.StackResources[] | select(.LogicalResourceId == "TestFunc")'`
FuncName=`echo $Resource | jq .PhysicalResourceId -r`
Region=`echo $Resource | jq .StackId -r | cut -d : -f 4`
AccountID=`echo $Resource | jq .StackId -r | cut -d : -f 5`

rm $OUTPUT_FILE
echo "PolicyLambdaArn=arn:aws:lambda:$Region:$AccountID:function:$FuncName"
