package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/event"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var errExpected = errors.New("expected")
var jsonNull = json.RawMessage("null")

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Error(msg string, v ...any) {
	m.Called(msg, v)
}

func (m *mockLogger) Info(msg string, v ...any) {
	m.Called(msg, v)
}

func (m *mockLogger) Debug(msg string, v ...any) {
	m.Called(msg, v)
}

type mockEventClient struct {
	mock.Mock
}

func (m *mockEventClient) SendLpaUpdated(ctx context.Context, event event.LpaUpdated) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

type mockStore struct {
	get    shared.Lpa
	getErr error
	put    any
	putErr error
	update shared.Update
}

func (m *mockStore) Get(context.Context, string) (shared.Lpa, error) { return m.get, m.getErr }
func (m *mockStore) Put(ctx context.Context, data any) error {
	m.put = data
	return m.putErr
}
func (m *mockStore) PutChanges(ctx context.Context, data any, update shared.Update) error {
	m.put = data
	m.update = update
	return m.putErr
}

type mockVerifier struct {
	claims shared.LpaStoreClaims
	err    error
}

func (m *mockVerifier) VerifyHeader(events.APIGatewayProxyRequest) (*shared.LpaStoreClaims, error) {
	return &m.claims, m.err
}

func TestHandleEvent(t *testing.T) {
	signedAt := time.Date(2022, time.January, 2, 12, 13, 14, 6, time.UTC)

	logger := &mockLogger{}
	logger.On("Debug", "Successfully parsed JWT from event header", mock.Anything)

	store := &mockStore{get: shared.Lpa{
		Uid: "1",
		LpaInit: shared.LpaInit{
			CertificateProvider: shared.CertificateProvider{
				Email:   "a@example.com",
				Channel: shared.ChannelPaper,
			},
		},
	}}

	client := mockEventClient{}
	client.On("SendLpaUpdated", mock.Anything, event.LpaUpdated{
		Uid:        "1",
		ChangeType: "CERTIFICATE_PROVIDER_SIGN",
	}).Return(nil)

	l := Lambda{
		eventClient: &client,
		store:       store,
		verifier: &mockVerifier{
			claims: shared.LpaStoreClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "1234",
				},
			},
		},
		logger: logger,
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"type":"CERTIFICATE_PROVIDER_SIGN","changes":[{"key":"/certificateProvider/signedAt","old":null,"new":"2022-01-02T12:13:14.000000006Z"},{"key":"/certificateProvider/contactLanguagePreference","old":null,"new":"en"},{"key":"/certificateProvider/email","old":"a@example.com","new":"b@example.com"},{"key":"/certificateProvider/channel","old":"paper","new":"online"}]}`,
	})

	assert.Nil(t, err)
	assert.Equal(t, 201, resp.StatusCode)
	assert.Contains(t, resp.Body, `"2022-01-02T12:13:14.000000006Z"`)
	assert.Contains(t, resp.Body, `"en"`)
	assert.Equal(t, shared.Lpa{
		Uid: "1",
		LpaInit: shared.LpaInit{
			CertificateProvider: shared.CertificateProvider{
				SignedAt:                  &signedAt,
				ContactLanguagePreference: shared.LangEn,
				Channel:                   shared.ChannelOnline,
				Email:                     "b@example.com",
			},
		},
	}, store.put)

	assert.NoError(t, uuid.Validate(store.update.Id))
	assert.Regexp(t, regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`), store.update.Applied)

	assert.True(t, cmp.Equal(
		shared.Update{
			Uid:    "1",
			Author: "1234",
			Type:   "CERTIFICATE_PROVIDER_SIGN",
			Changes: []shared.Change{
				{
					Key: "/certificateProvider/signedAt",
					Old: jsonNull,
					New: json.RawMessage(`"2022-01-02T12:13:14.000000006Z"`),
				},
				{
					Key: "/certificateProvider/contactLanguagePreference",
					Old: jsonNull,
					New: json.RawMessage(`"en"`),
				},
				{
					Key: "/certificateProvider/email",
					Old: json.RawMessage(`"a@example.com"`),
					New: json.RawMessage(`"b@example.com"`),
				},
				{
					Key: "/certificateProvider/channel",
					Old: json.RawMessage(`"paper"`),
					New: json.RawMessage(`"online"`),
				},
			},
		},
		store.update,
		cmpopts.IgnoreFields(shared.Update{}, "Id", "Applied"),
	))
	client.AssertExpectations(t)
}

func TestHandleEventWhenUnknownType(t *testing.T) {
	logger := &mockLogger{}
	logger.On("Debug", "Successfully parsed JWT from event header", mock.Anything)

	l := Lambda{
		store:    &mockStore{get: shared.Lpa{Uid: "1"}},
		verifier: &mockVerifier{},
		logger:   logger,
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"type":"SCANNING_CORRECTION","changes":[{"key":"/donor/firstNames","old":"Johm","new":"John"}]}`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	assert.JSONEq(t, `{"code":"INVALID_REQUEST","detail":"Invalid request","errors":[{"source":"/type","detail":"invalid value"}]}`, resp.Body)
}

func TestHandleEventWhenUpdateInvalid(t *testing.T) {
	logger := &mockLogger{}
	logger.On("Debug", "Successfully parsed JWT from event header", mock.Anything)

	l := Lambda{
		store:    &mockStore{get: shared.Lpa{Uid: "1"}},
		verifier: &mockVerifier{},
		logger:   logger,
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"type":"CERTIFICATE_PROVIDER_SIGN","changes":[]}`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	assert.JSONEq(t, `{"code":"INVALID_REQUEST","detail":"Invalid request","errors":[{"source":"/changes","detail":"missing /certificateProvider/signedAt"}]}`, resp.Body)
}

func TestHandleEventWhenLpaNotFound(t *testing.T) {
	logger := &mockLogger{}
	logger.On("Debug", "Successfully parsed JWT from event header", mock.Anything)
	logger.On("Debug", "Uid not found", mock.Anything)

	l := Lambda{
		store:    &mockStore{},
		verifier: &mockVerifier{},
		logger:   logger,
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{}`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode)
	assert.JSONEq(t, `{"code":"NOT_FOUND","detail":"Record not found"}`, resp.Body)
}

func TestHandleEventWhenStoreGetError(t *testing.T) {
	logger := &mockLogger{}
	logger.On("Debug", "Successfully parsed JWT from event header", mock.Anything)
	logger.On("Error", "error fetching LPA", []interface{}{slog.Any("err", errExpected)})

	l := Lambda{
		store:    &mockStore{getErr: errExpected},
		verifier: &mockVerifier{},
		logger:   logger,
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{}`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.JSONEq(t, `{"code":"INTERNAL_SERVER_ERROR","detail":"Internal server error"}`, resp.Body)
}

func TestHandleEventWhenRequestBodyNotJSON(t *testing.T) {
	logger := &mockLogger{}
	logger.On("Debug", "Successfully parsed JWT from event header", mock.Anything)
	logger.On("Error", "error unmarshalling request", mock.Anything)

	l := Lambda{
		verifier: &mockVerifier{},
		logger:   logger,
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.JSONEq(t, `{"code":"INTERNAL_SERVER_ERROR","detail":"Internal server error"}`, resp.Body)
}

func TestHandleEventWhenHeaderNotVerified(t *testing.T) {
	logger := &mockLogger{}
	logger.On("Info", "Unable to verify JWT from header", mock.Anything)

	l := Lambda{
		verifier: &mockVerifier{err: errors.New("Invalid JWT")},
		logger:   logger,
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 401, resp.StatusCode)
	assert.JSONEq(t, `{"code":"UNAUTHORISED","detail":"Invalid JWT"}`, resp.Body)
}

func TestHandleEventWhenSendLpaUpdatedFailed(t *testing.T) {
	store := &mockStore{get: shared.Lpa{Uid: "1"}}
	client := mockEventClient{}
	client.On("SendLpaUpdated", mock.Anything, mock.Anything).Return(errors.New("Update failed"))

	logger := mockLogger{}
	logger.On("Debug", "Successfully parsed JWT from event header", mock.Anything)
	logger.On("Error", "unexpected error occurred", []interface{}{slog.Any("err", errors.New("Update failed"))})

	l := Lambda{
		eventClient: &client,
		store:       store,
		verifier: &mockVerifier{
			claims: shared.LpaStoreClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "1234",
				},
			},
		},
		logger: &logger,
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"type":"CERTIFICATE_PROVIDER_SIGN","changes":[{"key":"/certificateProvider/signedAt","old":null,"new":"2022-01-02T12:13:14.000000006Z"},{"key":"/certificateProvider/contactLanguagePreference","old":null,"new":"en"},{"key":"/certificateProvider/email","old":null,"new":"b@example.com"}]}`,
	})

	client.AssertExpectations(t)
	logger.AssertExpectations(t)

	assert.Nil(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}
