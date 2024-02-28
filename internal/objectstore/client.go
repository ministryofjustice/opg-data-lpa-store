package objectstore

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type awsS3Client interface {
	PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, input *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type S3Client struct {
	bucketName string
	awsClient  awsS3Client
}

func (c *S3Client) Put(objectKey string, obj any) (*s3.PutObjectOutput, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return &s3.PutObjectOutput{}, err
	}

	return c.awsClient.PutObject(
		context.Background(),
		&s3.PutObjectInput{
			Bucket:               aws.String(c.bucketName),
			Key:                  aws.String(objectKey),
			Body:                 bytes.NewReader(b),
			ServerSideEncryption: types.ServerSideEncryptionAwsKms,
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

type resolver struct {
	URL string
}

func (r *resolver) ResolveEndpoint(service, region string) (aws.Endpoint, error) {
	return aws.Endpoint{ URL: r.URL, HostnameImmutable: true, }, nil
}

// set endpoint to "" outside dev to use default resolver
func NewS3Client(bucketName, endpointURL string) *S3Client {
	var (
		cfg aws.Config
		err error
	)

	if endpointURL == "" {
		cfg, err = config.LoadDefaultConfig(context.Background())
	} else {
		cfg, err = config.LoadDefaultConfig(
			context.Background(),
			config.WithEndpointResolver(&resolver{ URL: endpointURL }),
		)
	}

	if err != nil {
		panic(err)
	}

	awsClient := s3.NewFromConfig(cfg)

	return &S3Client{
		bucketName: bucketName,
		awsClient: awsClient,
	}
}
