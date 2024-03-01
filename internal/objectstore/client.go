package objectstore

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type awsS3Client interface {
	PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type S3Client struct {
	bucketName string
	awsClient  awsS3Client
}

func (c *S3Client) Put(objectKey string, obj any) error {
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = c.awsClient.PutObject(
		context.Background(),
		&s3.PutObjectInput{
			Bucket:               aws.String(c.bucketName),
			Key:                  aws.String(objectKey),
			Body:                 bytes.NewReader(b),
			ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		},
	)

	return err
}

// set endpoint to "" outside dev to use default resolver
func NewS3Client(awsConfig aws.Config, bucketName string) *S3Client {
	awsClient := s3.NewFromConfig(awsConfig)

	return &S3Client{
		bucketName: bucketName,
		awsClient:  awsClient,
	}
}
