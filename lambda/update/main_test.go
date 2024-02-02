package main

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/stretchr/testify/assert"
)

var expectedError = errors.New("expected")

type mockStore struct {
	get    shared.Lpa
	getErr error
	put    any
	putErr error
}

func (m *mockStore) Get(context.Context, string) (shared.Lpa, error) { return m.get, m.getErr }
func (m *mockStore) Put(ctx context.Context, data any) error {
	m.put = data
	return m.putErr
}
func (m *mockStore) PutChanges(ctx context.Context, data any, update shared.Update) error {
	m.put = data
	return m.putErr
}

type mockVerifier struct {
	claims shared.LpaStoreClaims
	err error
}

func (m *mockVerifier) VerifyHeader(events.APIGatewayProxyRequest) (*shared.LpaStoreClaims, error) {
	return &m.claims, m.err
}

func TestHandleEvent(t *testing.T) {
	store := &mockStore{get: shared.Lpa{Uid: "1"}}
	l := Lambda{
		store:    store,
		verifier: &mockVerifier{},
		logger:   logging.New(io.Discard, ""),
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"type":"CERTIFICATE_PROVIDER_SIGN","changes":[{"key":"/certificateProvider/signedAt","old":null,"new":"2022-01-02T12:13:14.000000006Z"},{"key":"/certificateProvider/contactLanguagePreference","old":null,"new":"en"}]}`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 201, resp.StatusCode)
	assert.Contains(t, resp.Body, `"2022-01-02T12:13:14.000000006Z"`)
	assert.Contains(t, resp.Body, `"en"`)
	assert.Equal(t, shared.Lpa{
		Uid: "1",
		LpaInit: shared.LpaInit{
			CertificateProvider: shared.CertificateProvider{
				SignedAt:                  time.Date(2022, time.January, 2, 12, 13, 14, 6, time.UTC),
				ContactLanguagePreference: shared.LangEn,
			},
		},
	}, store.put)
}

func TestHandleEventWhenUnknownType(t *testing.T) {
	l := Lambda{
		store:    &mockStore{get: shared.Lpa{Uid: "1"}},
		verifier: &mockVerifier{},
		logger:   logging.New(io.Discard, ""),
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"type":"SCANNING_CORRECTION","changes":[{"key":"/donor/firstNames","old":"Johm","new":"John"}]}`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	assert.JSONEq(t, `{"code":"INVALID_REQUEST","detail":"Invalid request","errors":[{"source":"/type","detail":"invalid value"}]}`, resp.Body)
}

func TestHandleEventWhenUpdateInvalid(t *testing.T) {
	l := Lambda{
		store:    &mockStore{get: shared.Lpa{Uid: "1"}},
		verifier: &mockVerifier{},
		logger:   logging.New(io.Discard, ""),
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"type":"CERTIFICATE_PROVIDER_SIGN","changes":[]}`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	assert.JSONEq(t, `{"code":"INVALID_REQUEST","detail":"Invalid request","errors":[{"source":"/changes","detail":"missing /certificateProvider/signedAt"},{"source":"/changes","detail":"missing /certificateProvider/contactLanguagePreference"}]}`, resp.Body)
}

func TestHandleEventWhenLpaNotFound(t *testing.T) {
	l := Lambda{
		store:    &mockStore{},
		verifier: &mockVerifier{},
		logger:   logging.New(io.Discard, ""),
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{}`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode)
	assert.JSONEq(t, `{"code":"NOT_FOUND","detail":"Record not found"}`, resp.Body)
}

func TestHandleEventWhenStoreGetError(t *testing.T) {
	l := Lambda{
		store:    &mockStore{getErr: expectedError},
		verifier: &mockVerifier{},
		logger:   logging.New(io.Discard, ""),
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{}`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.JSONEq(t, `{"code":"INTERNAL_SERVER_ERROR","detail":"Internal server error"}`, resp.Body)
}

func TestHandleEventWhenRequestBodyNotJSON(t *testing.T) {
	l := Lambda{
		verifier: &mockVerifier{},
		logger:   logging.New(io.Discard, ""),
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.JSONEq(t, `{"code":"INTERNAL_SERVER_ERROR","detail":"Internal server error"}`, resp.Body)
}

func TestHandleEventWhenHeaderNotVerified(t *testing.T) {
	l := Lambda{
		verifier: &mockVerifier{err: errors.New("Invalid JWT")},
		logger:   logging.New(io.Discard, ""),
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 401, resp.StatusCode)
	assert.JSONEq(t, `{"code":"UNAUTHORISED","detail":"Invalid JWT"}`, resp.Body)
}
