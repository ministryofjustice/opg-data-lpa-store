package objectstore

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
)

type Client interface {
	Put(name string, obj any) error
	Get(name string) (StoredObject, error)
}

type StoredObject struct {
	name string
}

type S3Client struct {
	awsClient *s3.Client
}

func (c *S3Client) Put(name string, obj any) error {
	/*c.awsClient.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})*/

	return nil
}

func (c *S3Client) Get(name string) (StoredObject, error) {
	return StoredObject{}, nil
}

func New(endpoint string) Client {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	// xray instrumentation
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)

	awsClient := s3.NewFromConfig(
		cfg,
		func (o *s3.Options) {
		    o.BaseEndpoint = aws.String(endpoint)
		},
	)

	return &S3Client{
		awsClient: awsClient,
	}
}
