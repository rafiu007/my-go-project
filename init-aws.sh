#!/bin/bash

# Wait for LocalStack to be ready
echo "Waiting for LocalStack to be ready..."
sleep 10

# Create the SQS queue
echo "Creating SQS queue..."
awslocal sqs create-queue \
    --queue-name calendar-entries

echo "Verifying queue creation..."
awslocal sqs list-queues

echo "SQS queue setup completed"
EOF
