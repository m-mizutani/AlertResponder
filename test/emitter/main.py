import json
import boto3
import logging
import os
import datetime


logger = logging.getLogger()
logger.setLevel(level=logging.INFO)


def iter_sns_msg(records):
    for record in records.get('Records', []):
        try:
            sns_msg = record.get('Sns', {}).get('Message')
            yield json.loads(sns_msg)
        except Exception as e:
            logger.error(e)


def handler(records, context):
    client = boto3.client('dynamodb')

    for msg in iter_sns_msg(records):
        logger.info(json.dumps(msg, indent=2))

        res = client.put_item(
            TableName=os.environ["TABLE_NAME"],
            Item={
                'report_id': {'S': msg['report_id']},
                'timestamp': {'N': str(datetime.datetime.utcnow().timestamp())},
                'report': {'B': json.dumps(msg)},
            })

        logger.info(res)

    return {'message': 'ok'}
