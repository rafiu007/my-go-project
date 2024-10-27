#!/bin/bash
# init-aws.sh

# Wait for LocalStack to be ready
sleep 5

# Create SQS queue
awslocal sqs create-queue \
    --queue-name calendar-entries \
    --attributes "{
        \"DelaySeconds\":\"0\",
        \"MessageRetentionPeriod\":\"86400\",
        \"VisibilityTimeout\":\"30\"
    }"

echo "SQS queue created successfully"