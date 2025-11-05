package objectstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

type awsS3Client interface {
	PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, input *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type presignClient interface {
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type S3Client struct {
	bucketName string
	awsClient  awsS3Client
	presigner  presignClient
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

func (c *S3Client) PresignLpa(ctx context.Context, lpa shared.Lpa) (shared.Lpa, error) {
	if len(lpa.RestrictionsAndConditionsImages) > 0 {
		for i, restrictionsImage := range lpa.RestrictionsAndConditionsImages {
			req, err := c.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
				Bucket: aws.String(c.bucketName),
				Key:    aws.String(restrictionsImage.Path),
			})
			if err != nil {
				return lpa, err
			}

			lpa.RestrictionsAndConditionsImages[i].Path = req.URL
		}
	}

	return lpa, nil
}

func (c *S3Client) Get(ctx context.Context, objectKey string) (string, error) {
	result, err := c.awsClient.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		return "", err
	}

	//nolint:errcheck
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)

	if err != nil {
		return "", err
	}

	return string(body), nil
}
