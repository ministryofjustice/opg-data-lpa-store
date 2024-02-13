#!/usr/bin/env bash

# S3
buckets=$(awslocal s3 ls)

echo $buckets | grep "opg-lpa-store-static-eu-west-1" || exit 1

# DynamoDB
tables=$(awslocal dynamodb list-tables)
echo $tables | grep '"deeds"' || exit 1
echo $tables | grep '"changes"' || exit 1
