#!/bin/bash

Resources=`aws cloudformation describe-stack-resources --stack-name $1 | jq '.StackResources[]'`
InspectorArn=`echo $Resources | jq 'select(.LogicalResourceId == "Inspector") | .PhysicalResourceId' -r`
ReporterArn=`echo $Resources | jq 'select(.LogicalResourceId == "Reporter") | .PhysicalResourceId' -r`

StackId=`echo $Resources | jq 'select(.LogicalResourceId == "Reporter") | .StackId' -r`
AccountId=`echo $StackId | cut -d : -f 5`
Region=`echo $StackId | cut -d : -f 4`

cat <<EOF > params.json
{
  "AccountId": "$AccountId",
  "Region": "$Region",
  "Inspector": "$InspectorArn",
  "Reporter": "$ReporterArn"
}
EOF
