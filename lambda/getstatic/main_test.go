package main

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	ctx        = context.WithValue(context.Background(), (*string)(nil), "testing")
	errExample = errors.New("err")
)

func TestLambdaHandleEventHeadersNotVerified(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"uid": "my-uid"},
	}

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(req).
		Return(nil, errExample)

	logger := newMockLogger(t)
	logger.EXPECT().
		Info("Unable to verify JWT from header")

	lambda := &Lambda{
		verifier: verifier,
		logger:   logger,
	}

	resp, err := lambda.HandleEvent(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: 401,
		Body:       `{"code":"UNAUTHORISED","detail":"Invalid JWT"}`,
	}, resp)
}

func TestLambdaHandleEventStaticLpaNotFound(t *testing.T) {
	testcases := map[string]map[string]any{
		"generic error, 500 response": {
			"error":                errExample,
			"expectedResponseCode": 500,
			"expectedResponseBody": `{"code":"INTERNAL_SERVER_ERROR","detail":"Internal server error"}`,
		},
		"NoSuchUpload error, 404 response": {
			"error":                &types.NoSuchUpload{},
			"expectedResponseCode": 404,
			"expectedResponseBody": `{"code":"NOT_FOUND","detail":"Record not found"}`,
		},
		"NoSuchKey error, 404 response": {
			"error":                &types.NoSuchKey{},
			"expectedResponseCode": 404,
			"expectedResponseBody": `{"code":"NOT_FOUND","detail":"Record not found"}`,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				PathParameters: map[string]string{"uid": "my-uid"},
			}

			verifier := newMockVerifier(t)
			verifier.EXPECT().
				VerifyHeader(req).
				Return(nil, nil)

			logger := newMockLogger(t)
			logger.EXPECT().
				Debug("Successfully parsed JWT from event header")

			staticLpaStorage := newMockS3Client(t)
			staticLpaStorage.EXPECT().
				Get(ctx, "my-uid/donor-executed-lpa.json").
				Return("", tc["error"].(error))

			logger.EXPECT().
				Error("error fetching static LPA", mock.Anything)

			lambda := &Lambda{
				verifier:         verifier,
				logger:           logger,
				staticLpaStorage: staticLpaStorage,
			}

			resp, err := lambda.HandleEvent(ctx, req)
			assert.Nil(t, err)
			assert.Equal(t, events.APIGatewayProxyResponse{
				StatusCode: tc["expectedResponseCode"].(int),
				Body:       tc["expectedResponseBody"].(string),
			}, resp)
		})
	}
}

func TestLambdaHandleEventStaticLpaReturned(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"uid": "my-uid"},
	}

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(req).
		Return(nil, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header")

	staticLpaStorage := newMockS3Client(t)
	staticLpaStorage.EXPECT().
		Get(ctx, "my-uid/donor-executed-lpa.json").
		Return("Static LPA data", nil)

	lambda := &Lambda{
		verifier:         verifier,
		logger:           logger,
		staticLpaStorage: staticLpaStorage,
	}

	resp, err := lambda.HandleEvent(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Static LPA data",
	}, resp)
}
