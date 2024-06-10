package objectstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

type awsS3Client interface {
	PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type S3Client struct {
	bucketName string
	awsClient  awsS3Client
}

func (c *S3Client) Put(ctx context.Context, objectKey string, obj any) error {
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = c.awsClient.PutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:               aws.String(c.bucketName),
			Key:                  aws.String(objectKey),
			Body:                 bytes.NewReader(b),
			ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		},
	)

	return err
}

func (c *S3Client) UploadFile(ctx context.Context, file shared.FileUpload, path string) (shared.File, error) {
	var imgData []byte
	var err error

	if imgData, err = base64.StdEncoding.DecodeString(file.Data); err != nil {
		return shared.File{}, err
	}

	_, err = c.awsClient.PutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:               aws.String(c.bucketName),
			Key:                  aws.String(path),
			Body:                 bytes.NewReader(imgData),
			ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		},
	)

	if err != nil {
		return shared.File{}, err
	}

	hash := sha256.New()
	_, err = hash.Write(imgData)

	if err != nil {
		return shared.File{}, err
	}

	return shared.File{
		Path: path,
		Hash: hex.EncodeToString(hash.Sum(nil)),
	}, nil

}

type resolver struct {
	endpointURL string
}

func (r *resolver) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (
	smithyendpoints.Endpoint, error,
) {
	u, err := url.Parse(r.endpointURL)
	if err != nil {
		return smithyendpoints.Endpoint{}, err
	}

	u.Path = *params.Bucket

	return smithyendpoints.Endpoint{
		URI: *u,
	}, nil
}

func NewS3Client(awsConfig aws.Config, endpointURL string, bucketName string) *S3Client {
	awsClient := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		if endpointURL != "" {
			o.EndpointResolverV2 = &resolver{
				endpointURL: endpointURL,
			}
		}
	})

	return &S3Client{
		bucketName: bucketName,
		awsClient:  awsClient,
	}
}
