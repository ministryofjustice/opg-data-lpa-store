package getstatic

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
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
		Get(ctx, "my-uid").
		Return("", errExample)

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
		StatusCode: 500,
		Body:       `{"code":"INTERNAL_SERVER_ERROR","detail":"Internal server error"}`,
	}, resp)
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
		Get(ctx, "my-uid").
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
