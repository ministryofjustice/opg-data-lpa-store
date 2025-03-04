package objectstore

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	ctx           = context.WithValue(context.Background(), "for", "testing")
	expectedError = errors.New("expected")
	bucketName    = "a-bucket"
	objectKey     = "an-object-key"
)

func TestS3ClientPut(t *testing.T) {
	awsS3Client := newMockAwsS3Client(t)
	awsS3Client.EXPECT().
		PutObject(ctx, &s3.PutObjectInput{
			Bucket:               aws.String(bucketName),
			Key:                  aws.String(objectKey),
			Body:                 bytes.NewReader([]byte(`{"ID":1}`)),
			ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		}).
		Return(nil, expectedError)

	client := &S3Client{
		bucketName: bucketName,
		awsClient:  awsS3Client,
	}

	err := client.Put(ctx, objectKey, struct{ ID int }{ID: 1})
	assert.Equal(t, expectedError, err)
}

func TestS3ClientUploadFile(t *testing.T) {
	upload := shared.FileUpload{
		Filename: "myfile.txt",
		Data:     "Q29udGVudHMgb2YgbXkgZmlsZQ==",
	}

	awsS3Client := newMockAwsS3Client(t)
	awsS3Client.EXPECT().
		PutObject(ctx, &s3.PutObjectInput{
			Bucket:               aws.String(bucketName),
			Key:                  aws.String("dir/myfile.txt"),
			Body:                 bytes.NewReader([]byte("Contents of my file")),
			ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		}).
		Return(&s3.PutObjectOutput{}, nil)

	client := &S3Client{
		bucketName: bucketName,
		awsClient:  awsS3Client,
	}

	file, err := client.UploadFile(ctx, upload, "dir/myfile.txt")
	assert.Nil(t, err)
	assert.Equal(t, "dir/myfile.txt", file.Path)
	assert.Equal(t, "7ac4f2b48096ac5f4600a0775563d0f2b3369a3ea00d1fa813f45c18620dba28", file.Hash)
}

func TestS3ClientUploadFileWhenDecodingError(t *testing.T) {
	upload := shared.FileUpload{
		Filename: "myfile.txt",
		Data:     "This isn't base 64",
	}

	client := &S3Client{}

	_, err := client.UploadFile(ctx, upload, "dir/myfile.txt")
	assert.ErrorContains(t, err, "illegal base64 data")
}

func TestS3ClientUploadFileWhenS3Error(t *testing.T) {
	upload := shared.FileUpload{
		Filename: "myfile.txt",
		Data:     "Q29udGVudHMgb2YgbXkgZmlsZQ==",
	}

	awsS3Client := newMockAwsS3Client(t)
	awsS3Client.EXPECT().
		PutObject(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	client := &S3Client{
		bucketName: "bucket1",
		awsClient:  awsS3Client,
	}

	_, err := client.UploadFile(ctx, upload, "dir/myfile.txt")
	assert.Equal(t, expectedError, err)
}

func TestS3ClientPresignLpa(t *testing.T) {
	presigner := newMockPresignClient(t)
	presigner.EXPECT().
		PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String("bucket1"),
			Key:    aws.String("x.jpg"),
		}).
		Return(&v4.PresignedHTTPRequest{URL: "aws/x.jpg?blah"}, nil).
		Once()
	presigner.EXPECT().
		PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String("bucket1"),
			Key:    aws.String("y.png"),
		}).
		Return(&v4.PresignedHTTPRequest{URL: "aws/y.png?blah"}, nil).
		Once()

	client := &S3Client{
		bucketName: "bucket1",
		presigner:  presigner,
	}

	result, err := client.PresignLpa(ctx, shared.Lpa{
		RestrictionsAndConditionsImages: []shared.File{
			{
				Path: "x.jpg",
				Hash: "xxx",
			},
			{
				Path: "y.png",
				Hash: "yyy",
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "aws/x.jpg?blah", result.RestrictionsAndConditionsImages[0].Path)
	assert.Equal(t, "aws/y.png?blah", result.RestrictionsAndConditionsImages[1].Path)
}

func TestS3ClientPresignLpaWhenError(t *testing.T) {
	presigner := newMockPresignClient(t)
	presigner.EXPECT().
		PresignGetObject(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	client := &S3Client{
		bucketName: "bucket1",
		presigner:  presigner,
	}

	_, err := client.PresignLpa(ctx, shared.Lpa{
		RestrictionsAndConditionsImages: []shared.File{{Path: "x.jpg", Hash: "xxx"}},
	})
	assert.ErrorIs(t, err, expectedError)
}
