#!/bin/bash

TEMPLATE_FILE=template.yml
OUTPUT_FILE=`mktemp`

if [ $# -le 3 ]; then
    echo "usage) $0 AlertResponderStackName StackName CodeS3Bucket CodeS3Prefix [LambdaArn]"
    exit 1
fi

if [ "$5" != "" ]; then
   OPT="LambdaArn=$2"
else
   OPT=""
fi

echo "Output Template: $OUTPUT_FILE"
echo ""

ResponderResources=`aws cloudformation describe-stack-resources --stack-name $1 | jq '.StackResources[]'`

TaskStream=`echo $ResponderResources | jq 'select(.LogicalResourceId == "TaskStream")'`

TaskStreamName=`echo $TaskStream | jq .PhysicalResourceId -r`
Region=`echo $TaskStream | jq .StackId -r | cut -d : -f 4`
AccountID=`echo $TaskStream | jq .StackId -r | cut -d : -f 5`

ReportDataName=`echo $ResponderResources | jq 'select(.LogicalResourceId == "ReportData") | .PhysicalResourceId' -r`

TaskStreamArn="arn:aws:kinesis:$Region:$AccountID:stream/$TaskStreamName"
ReportLineArn=`echo $ResponderResources | jq 'select(.LogicalResourceId == "ReportLine") | .PhysicalResourceId' -r`
IncidentLineArn=`echo $ResponderResources | jq 'select(.LogicalResourceId == "IncidentLine") | .PhysicalResourceId' -r`


# aws cloudformation package --template-file $TEMPLATE_FILE \
#     --output-template-file $OUTPUT_FILE --s3-bucket $3 --s3-prefix $4
# aws cloudformation deploy --template-file $OUTPUT_FILE --stack-name $2 \
#     --capabilities CAPABILITY_IAM \
#     --parameter-overrides $OPT \
#     ReportLineArn=$ReportLineArn \
#     IncidentLineArn=$IncidentLineArn \
#     TaskStreamArn=$TaskStreamArn \
#     ReportDataName=$ReportDataName 
# 
Resources=`aws cloudformation describe-stack-resources --stack-name $2 | jq '.StackResources[]'`

ReportReceiverArn=`echo $Resources | jq 'select(.LogicalResourceId == "ReportReceiver") | .PhysicalResourceId' -r`
IncidentReceiverArn=`echo $Resources | jq 'select(.LogicalResourceId == "IncidentReceiver") | .PhysicalResourceId' -r`

echo $Resources | jq .
echo $ReportReceiverArn
echo $IncidentReceiverArn

cat <<EOF > test.json
{
  "StackName": "$1",
  "Region": "$Region",
  "ReportReceiverArn": "$ReportReceiverArn",
  "IncidentReceiverArn": "$IncidentReceiverArn"
}
EOF

echo ""
echo "done"

