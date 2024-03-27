#! /usr/bin/env bash
create_bucket() {
    BUCKET=$1
    # Create Private Bucket
    awslocal s3api create-bucket \
        --acl private \
        --region eu-west-1 \
        --create-bucket-configuration LocationConstraint=eu-west-1 \
        --bucket "$BUCKET"

    # Add Public Access Block
    awslocal s3api put-public-access-block \
        --public-access-block-configuration "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true" \
        --bucket "$BUCKET"

    # Add Default Encryption
    awslocal s3api put-bucket-encryption \
        --bucket "$BUCKET" \
        --server-side-encryption-configuration '{ "Rules": [ { "ApplyServerSideEncryptionByDefault": { "SSEAlgorithm": "AES256" } } ] }'

    # Add Encryption Policy
    awslocal s3api put-bucket-policy \
        --policy '{ "Statement": [ { "Sid": "DenyUnEncryptedObjectUploads", "Effect": "Deny", "Principal": { "AWS": "*" }, "Action": "s3:PutObject", "Resource": "arn:aws:s3:::'${BUCKET}'/*", "Condition":  { "StringNotEquals": { "s3:x-amz-server-side-encryption": "AES256" } } }, { "Sid": "DenyUnEncryptedObjectUploads", "Effect": "Deny", "Principal": { "AWS": "*" }, "Action": "s3:PutObject", "Resource": "arn:aws:s3:::'${BUCKET}'/*", "Condition":  { "Bool": { "aws:SecureTransport": false } } } ] }' \
        --bucket "$BUCKET"

    # Add Bucket Versioning
    awslocal s3api put-bucket-versioning \
        --versioning-configuration '{ "MFADelete": "Disabled", "Status": "Enabled" }' \
        --bucket "$BUCKET"
}

# S3
create_bucket "opg-lpa-store-static-eu-west-1"

# DynamoDB
awslocal dynamodb create-table \
    --table-name deeds \
    --attribute-definitions AttributeName=uid,AttributeType=S \
    --key-schema AttributeName=uid,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST

awslocal dynamodb create-table \
    --table-name changes \
    --attribute-definitions AttributeName=uid,AttributeType=S AttributeName=applied,AttributeType=S \
    --key-schema AttributeName=uid,KeyType=HASH AttributeName=applied,KeyType=RANGE \
    --billing-mode PAY_PER_REQUEST

# Secrets Manager
awslocal secretsmanager create-secret --name local/jwt-key \
    --description "JWT secret for service authentication" \
    --secret-string "secret"
