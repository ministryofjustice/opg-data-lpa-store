package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var (
	ctx           = context.WithValue(context.Background(), (*string)(nil), "testing")
	expectedError = errors.New("err")
)

func TestLambdaHandleEvent(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"uid": "my-uid"},
	}

	lpa := shared.Lpa{Uid: "my-uid"}
	body, _ := json.Marshal(lpa)

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(req).
		Return(nil, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header")

	store := newMockStore(t)
	store.EXPECT().
		Get(ctx, "my-uid").
		Return(lpa, nil)

	lambda := &Lambda{
		verifier: verifier,
		logger:   logger,
		store:    store,
	}

	resp, err := lambda.HandleEvent(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, resp)
}

func TestLambdaHandleEventWhenPresignImages(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		PathParameters:        map[string]string{"uid": "my-uid"},
		QueryStringParameters: map[string]string{"presign-images": ""},
	}

	lpa := shared.Lpa{Uid: "my-uid"}
	presignedLpa := shared.Lpa{Uid: "my-uid2"}
	body, _ := json.Marshal(presignedLpa)

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(req).
		Return(nil, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header")

	store := newMockStore(t)
	store.EXPECT().
		Get(ctx, "my-uid").
		Return(lpa, nil)

	presignClient := newMockPresignClient(t)
	presignClient.EXPECT().
		PresignLpa(ctx, lpa).
		Return(presignedLpa, nil)

	lambda := &Lambda{
		verifier:      verifier,
		presignClient: presignClient,
		logger:        logger,
		store:         store,
	}

	resp, err := lambda.HandleEvent(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, resp)
}

func TestLambdaHandleEventWhenPresignImagesErrors(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		PathParameters:        map[string]string{"uid": "my-uid"},
		QueryStringParameters: map[string]string{"presign-images": ""},
	}

	lpa := shared.Lpa{Uid: "my-uid"}

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(req).
		Return(nil, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header")
	logger.EXPECT().
		Error("error signing URL", mock.Anything)

	store := newMockStore(t)
	store.EXPECT().
		Get(ctx, "my-uid").
		Return(lpa, nil)

	presignClient := newMockPresignClient(t)
	presignClient.EXPECT().
		PresignLpa(ctx, lpa).
		Return(shared.Lpa{}, expectedError)

	lambda := &Lambda{
		verifier:      verifier,
		presignClient: presignClient,
		logger:        logger,
		store:         store,
	}

	resp, err := lambda.HandleEvent(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       `{"code":"INTERNAL_SERVER_ERROR","detail":"Internal server error"}`,
	}, resp)
}

func TestLambdaHandleEventWhenUnauthorised(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"uid": "my-uid"},
	}

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(req).
		Return(nil, expectedError)

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

func TestLambdaHandleEventWhenNotFound(t *testing.T) {
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
	logger.EXPECT().
		Debug("Uid not found")

	store := newMockStore(t)
	store.EXPECT().
		Get(ctx, "my-uid").
		Return(shared.Lpa{}, nil)

	lambda := &Lambda{
		verifier: verifier,
		logger:   logger,
		store:    store,
	}

	resp, err := lambda.HandleEvent(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: 404,
		Body:       `{"code":"NOT_FOUND","detail":"Record not found"}`,
	}, resp)
}

func TestLambdaHandleEventWhenStoreErrors(t *testing.T) {
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
	logger.EXPECT().
		Error("error fetching LPA", slog.Any("err", expectedError))

	store := newMockStore(t)
	store.EXPECT().
		Get(ctx, "my-uid").
		Return(shared.Lpa{}, expectedError)

	lambda := &Lambda{
		verifier: verifier,
		logger:   logger,
		store:    store,
	}

	resp, err := lambda.HandleEvent(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       `{"code":"INTERNAL_SERVER_ERROR","detail":"Internal server error"}`,
	}, resp)
}
