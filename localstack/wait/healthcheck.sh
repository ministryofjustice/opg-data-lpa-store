#!/usr/bin/env bash

# S3
buckets=$(awslocal s3 ls)

echo $buckets | grep "opg-lpa-store-static-eu-west-1" || exit 1
