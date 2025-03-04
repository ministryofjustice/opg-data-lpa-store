package objectstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

type awsS3Client interface {
	PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type S3Client struct {
	bucketName string
	awsClient  awsS3Client
	presigner  *s3.PresignClient
}

func NewS3Client(awsConfig aws.Config, bucketName string) *S3Client {
	awsClient := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Client{
		bucketName: bucketName,
		awsClient:  awsClient,
		presigner:  s3.NewPresignClient(awsClient),
	}
}

func (c *S3Client) Put(ctx context.Context, objectKey string, obj any) error {
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = c.awsClient.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               aws.String(c.bucketName),
		Key:                  aws.String(objectKey),
		Body:                 bytes.NewReader(b),
		ServerSideEncryption: types.ServerSideEncryptionAwsKms,
	})

	return err
}

func (c *S3Client) UploadFile(ctx context.Context, file shared.FileUpload, path string) (shared.File, error) {
	imgData, err := base64.StdEncoding.DecodeString(file.Data)
	if err != nil {
		return shared.File{}, err
	}

	if _, err := c.awsClient.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               aws.String(c.bucketName),
		Key:                  aws.String(path),
		Body:                 bytes.NewReader(imgData),
		ServerSideEncryption: types.ServerSideEncryptionAwsKms,
	}); err != nil {
		return shared.File{}, err
	}

	hash := sha256.New()
	if _, err := hash.Write(imgData); err != nil {
		return shared.File{}, err
	}

	return shared.File{
		Path: path,
		Hash: hex.EncodeToString(hash.Sum(nil)),
	}, nil
}

func (c *S3Client) Presign(ctx context.Context, path string) (string, error) {
	req, err := c.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		return "", err
	}

	return req.URL, nil
}
