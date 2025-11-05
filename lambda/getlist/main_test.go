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
	"github.com/stretchr/testify/mock"
)

type ctxValueType string

const ctxValue ctxValueType = "for"

var (
	ctx         = context.WithValue(context.Background(), ctxValue, "testing")
	errExpected = errors.New("expect")
)

func TestLambdaHandleEvent(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		Body: `{"uids":["my-uid","another-uid"]}`,
	}

	lpas := []shared.Lpa{{Uid: "my-uid"}, {Uid: "another-uid"}}
	body, _ := json.Marshal(lpasResponse{Lpas: lpas})

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(req).
		Return(nil, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header")

	store := newMockStore(t)
	store.EXPECT().
		GetList(ctx, []string{"my-uid", "another-uid"}).
		Return(lpas, nil)

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
		Body:                  `{"uids":["my-uid","another-uid"]}`,
		QueryStringParameters: map[string]string{"presign-images": ""},
	}

	lpas := []shared.Lpa{{Uid: "my-uid"}, {Uid: "another-uid"}}
	presignedLpas := []shared.Lpa{{Uid: "my-uid2"}, {Uid: "another-uid2"}}
	body, _ := json.Marshal(lpasResponse{Lpas: presignedLpas})

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(req).
		Return(nil, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header")

	store := newMockStore(t)
	store.EXPECT().
		GetList(ctx, []string{"my-uid", "another-uid"}).
		Return(lpas, nil)

	presignClient := newMockPresignClient(t)
	presignClient.EXPECT().
		PresignLpa(ctx, lpas[0]).
		Return(presignedLpas[0], nil).
		Once()
	presignClient.EXPECT().
		PresignLpa(ctx, lpas[1]).
		Return(presignedLpas[1], nil).
		Once()

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
		Body:                  `{"uids":["my-uid","another-uid"]}`,
		QueryStringParameters: map[string]string{"presign-images": ""},
	}

	lpas := []shared.Lpa{{Uid: "my-uid"}, {Uid: "another-uid"}}

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
		GetList(ctx, []string{"my-uid", "another-uid"}).
		Return(lpas, nil)

	presignClient := newMockPresignClient(t)
	presignClient.EXPECT().
		PresignLpa(mock.Anything, mock.Anything).
		Return(shared.Lpa{}, errExpected)

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
		Body: `{"uids":["my-uid","another-uid"]}`,
	}

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(req).
		Return(nil, errors.New("hey"))

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

func TestLambdaHandleEventWhenBadRequest(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		Body: `{`,
	}

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(req).
		Return(nil, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header")
	logger.EXPECT().
		Error("error unmarshalling request", mock.Anything)

	lambda := &Lambda{
		verifier: verifier,
		logger:   logger,
	}

	resp, err := lambda.HandleEvent(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       `{"code":"INTERNAL_SERVER_ERROR","detail":"Internal server error"}`,
	}, resp)
}

func TestLambdaHandleEventWhenStoreErrors(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		Body: `{"uids":["my-uid","another-uid"]}`,
	}

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(req).
		Return(nil, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header")
	logger.EXPECT().
		Error("error fetching LPAs", slog.Any("err", errExpected))

	store := newMockStore(t)
	store.EXPECT().
		GetList(ctx, []string{"my-uid", "another-uid"}).
		Return(nil, errExpected)

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
