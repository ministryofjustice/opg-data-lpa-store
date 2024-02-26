package objectstore

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testCase struct {
	description string
	mockClientGetError error
	expectPut bool
	mockClientPutError error
	expectedReturnError error
}

type mockS3Client struct {
	mock.Mock
}

func (m *mockS3Client) Put(objectKey string, obj any) (*s3.PutObjectOutput, error) {
	args := m.Called(objectKey, obj)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *mockS3Client) Get(objectKey string) (*s3.GetObjectOutput, error) {
	args := m.Called(objectKey)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func TestSave(t *testing.T) {
	lpa := shared.Lpa{ Uid: "M-QQQQ-EEEE-TTTT" }
	expectedObjectKey := fmt.Sprintf("%s/donor-executed-lpa.json", lpa.Uid)

	testCases := []testCase{
		testCase{
			description: "FAIL - object already exists in S3 and shouldn't be overwritten",
			mockClientGetError: nil,
			expectPut: false,
			mockClientPutError: nil,
			expectedReturnError: fmt.Errorf("Could not save donor executed LPA as key %s already exists", expectedObjectKey),
		},

		testCase{
			description: "FAIL - S3 get of object key fails due to HTTP error",
			mockClientGetError: errors.New("HTTP transport error"),
			expectPut: false,
			mockClientPutError: nil,
			expectedReturnError: errors.New("HTTP transport error"),
		},

		testCase{
			description: "FAIL - S3 put of object key fails due to HTTP error",
			mockClientGetError: &types.NoSuchKey{},
			expectPut: true,
			mockClientPutError: errors.New("HTTP transport error"),
			expectedReturnError: errors.New("HTTP transport error"),
		},

		testCase{
			description: "SUCCESS - object does not exist in S3, and put of object to S3 is OK",
			mockClientGetError: &types.NoSuchKey{},
			expectPut: true,
			mockClientPutError: nil,
			expectedReturnError: nil,
		},
	}

	for _, tc := range(testCases) {
		t.Run(tc.description, func (t *testing.T) {
			mockClient := mockS3Client{}
			mockClient.On("Get", expectedObjectKey).Return(&s3.GetObjectOutput{}, tc.mockClientGetError)

			if tc.expectPut {
				mockClient.On("Put", expectedObjectKey, mock.Anything).Return(&s3.PutObjectOutput{}, tc.mockClientPutError)
			}

			sut := NewStaticLpaStorage(&mockClient)
			actualErr := sut.Save(&lpa)

			assert.Equal(t, tc.expectedReturnError, actualErr)
			mockClient.AssertExpectations(t)
		})
	}
}
