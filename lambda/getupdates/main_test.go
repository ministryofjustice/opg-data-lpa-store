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
)

var (
	ctx        = context.WithValue(context.Background(), (*string)(nil), "testing")
	errExample = errors.New("err")
)

func TestLambdaHandleEvent(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"uid": "my-uid"},
	}

	updates := []shared.Update{
		{
			Uid: "my-uid",
		},
	}
	body, _ := json.Marshal(updates)

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(req).
		Return(nil, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header")

	store := newMockStore(t)
	store.EXPECT().
		GetChanges(ctx, "my-uid").
		Return(updates, nil)

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

func TestLambdaHandleEventWhenUnauthorised(t *testing.T) {
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

func TestLambdaHandleEventWhenNoChangesFound(t *testing.T) {
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
		Debug("No updates found")

	store := newMockStore(t)
	store.EXPECT().
		GetChanges(ctx, "my-uid").
		Return([]shared.Update{}, nil)

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
		Error("error fetching updates", slog.Any("err", errExample))

	store := newMockStore(t)
	store.EXPECT().
		GetChanges(ctx, "my-uid").
		Return(nil, errExample)

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
