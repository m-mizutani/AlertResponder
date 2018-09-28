#!/bin/bash

REGIONS=(`cat scripts/cloudformation-regions.txt | tr '\n' ' '`)

for region in ${REGIONS[@]}; do
    aws --region $region cloudformation package \
                --template-file template.yml \
                --s3-bucket cfn-assets.$region \
                --s3-prefix AlertResponder/functions \
                --output-template-file sam.$region.yml
    aws s3 cp sam.$region.yml s3://cfn-assets.$region/AlertResponder/templates/latest.yml
done
