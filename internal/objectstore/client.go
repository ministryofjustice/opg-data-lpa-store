package objectstore

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
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

type resolverV2 struct {
	URL string
}

func (r *resolverV2) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (smithyendpoints.Endpoint, error) {
	if r.URL != "" {
		u, err := url.Parse(r.URL)
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{ URI: *u }, nil
	}

	return s3.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}

// set endpoint to "" outside dev to use default resolver
func NewS3Client(bucketName, endpointURL string) *S3Client {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	awsClient := s3.NewFromConfig(cfg, func (o *s3.Options) {
		o.EndpointResolverV2 = &resolverV2{
			URL: endpointURL,
		}
	})

	return &S3Client{
		bucketName: bucketName,
		awsClient: awsClient,
	}
}
