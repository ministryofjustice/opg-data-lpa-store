package objectstore

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client interface {
	Put(objectKey string, obj any) (*s3.PutObjectOutput, error)
	Get(objectKey string) (*s3.GetObjectOutput, error)
}

type S3Client struct {
	bucketName string
	awsClient  *s3.Client
}

func (c *S3Client) Put(objectKey string, obj any) (*s3.PutObjectOutput, error) {
    b, err := json.Marshal(obj)
    if err != nil {
        return &s3.PutObjectOutput{}, err
    }

	// first return value is output
	return c.awsClient.PutObject(
		context.Background(),
		&s3.PutObjectInput{
			Bucket: aws.String(c.bucketName),
			Key:    aws.String(objectKey),
			Body:   bytes.NewReader(b),
		},
	)
}

func (c *S3Client) Get(objectKey string) (*s3.GetObjectOutput, error) {
	return c.awsClient.GetObject(
		context.Background(),
		&s3.GetObjectInput{
			Bucket: aws.String(c.bucketName),
			Key:    aws.String(objectKey),
		},
	)
}

type endpointResolver struct {
	URL string
}

func (er *endpointResolver) ResolveEndpoint(service, region string) (aws.Endpoint, error) {
	return aws.Endpoint{ URL: er.URL, HostnameImmutable: true, }, nil
}

// set endpoint to "" outside dev to use default resolver
func New(endpointURL string) Client {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithEndpointResolver(&endpointResolver{ URL: endpointURL }),
	)

	if err != nil {
		panic(err)
	}

	awsClient := s3.NewFromConfig(cfg)

	return &S3Client{
		bucketName: "opg-lpa-store-static-eu-west-1",
		awsClient: awsClient,
	}
}
