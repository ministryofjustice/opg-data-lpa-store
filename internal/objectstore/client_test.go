package objectstore

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAwsClient struct {
	mock.Mock
}

func (m *mockAwsClient) PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func TestPut(t *testing.T) {
	client := mockAwsClient{}
	client.On("PutObject", mock.Anything, mock.Anything, mock.Anything).Return(&s3.PutObjectOutput{}, nil)

	c := S3Client{
		bucketName: "bucket1",
		awsClient:  &client,
	}

	err := c.Put(context.Background(), "anobject", struct{ ID int }{ID: 1})

	assert.Equal(t, nil, err)
	client.AssertExpectations(t)
}

func TestUploadFile(t *testing.T) {
	upload := shared.FileUpload{
		Filename: "myfile.txt",
		Data:     "Q29udGVudHMgb2YgbXkgZmlsZQ==",
	}

	client := mockAwsClient{}
	client.On("PutObject", mock.Anything, &s3.PutObjectInput{
		Bucket:               aws.String("bucket1"),
		Key:                  aws.String("dir/myfile.txt"),
		Body:                 bytes.NewReader([]byte("Contents of my file")),
		ServerSideEncryption: types.ServerSideEncryptionAwsKms,
	}).Return(&s3.PutObjectOutput{}, nil)

	c := S3Client{
		bucketName: "bucket1",
		awsClient:  &client,
	}

	file, err := c.UploadFile(context.Background(), upload, "dir/myfile.txt")

	assert.Nil(t, err)
	assert.Equal(t, "dir/myfile.txt", file.Path)
	assert.Equal(t, "bad0c316dc914dc22793a27828fc3064f057db42", file.Hash)
	client.AssertExpectations(t)
}

func TestUploadFileDecodingError(t *testing.T) {
	upload := shared.FileUpload{
		Filename: "myfile.txt",
		Data:     "This isn't base 64",
	}

	c := S3Client{
		bucketName: "bucket1",
		awsClient:  &mockAwsClient{},
	}

	file, err := c.UploadFile(context.Background(), upload, "dir/myfile.txt")

	assert.Equal(t, file, shared.File{})
	assert.Contains(t, err.Error(), "illegal base64 data")
}

func TestUploadFileS3Error(t *testing.T) {
	upload := shared.FileUpload{
		Filename: "myfile.txt",
		Data:     "Q29udGVudHMgb2YgbXkgZmlsZQ==",
	}

	expectedErr := errors.New("could not save object")

	client := mockAwsClient{}
	client.On("PutObject", mock.Anything, &s3.PutObjectInput{
		Bucket:               aws.String("bucket1"),
		Key:                  aws.String("dir/myfile.txt"),
		Body:                 bytes.NewReader([]byte("Contents of my file")),
		ServerSideEncryption: types.ServerSideEncryptionAwsKms,
	}).Return(&s3.PutObjectOutput{}, expectedErr)

	c := S3Client{
		bucketName: "bucket1",
		awsClient:  &client,
	}

	file, err := c.UploadFile(context.Background(), upload, "dir/myfile.txt")

	assert.Equal(t, file, shared.File{})
	assert.Equal(t, expectedErr, err)
	client.AssertExpectations(t)
}
