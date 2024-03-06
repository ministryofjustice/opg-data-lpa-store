package objectstore

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
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