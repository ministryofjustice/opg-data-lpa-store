package main

import (
	"bytes"
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

type mockVerifier struct{ ok bool }

func (m *mockVerifier) VerifyHeader(events.APIGatewayProxyRequest) bool { return m.ok }

func TestHandleEvent(t *testing.T) {
	now := time.Now()
	store := &mockStore{get: shared.Lpa{Uid: "1"}}
	l := Lambda{
		now:      func() time.Time { return now },
		store:    store,
		verifier: &mockVerifier{ok: true},
		logger:   logging.New(io.Discard, ""),
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"address":{"line1":"x","town":"y","country":"ZZ"},"signedAt":"2022-01-02T12:13:14.000000006Z","contactLanguagePreference":"en"}`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 201, resp.StatusCode)
	assert.JSONEq(t, `{}`, resp.Body)
	assert.Equal(t, CertificateProvider{
		UpdatedAt:                 now,
		Address:                   shared.Address{Line1: "x", Town: "y", Country: "ZZ"},
		SignedAt:                  time.Date(2022, time.January, 2, 12, 13, 14, 6, time.UTC),
		ContactLanguagePreference: shared.LangEn,
	}, store.put)
}

func TestHandleEventWhenPutErrors(t *testing.T) {
	var buf bytes.Buffer
	l := Lambda{
		now:      time.Now,
		store:    &mockStore{get: shared.Lpa{Uid: "1"}, putErr: expectedError},
		verifier: &mockVerifier{ok: true},
		logger:   logging.New(&buf, ""),
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"signedAt":"2022-01-02T12:13:14.000006Z","contactLanguagePreference":"en"}`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.JSONEq(t, `{"code":"INTERNAL_SERVER_ERROR","detail":"Internal server error"}`, resp.Body)
	assert.Contains(t, buf.String(), "expected")
}

func TestHandleEventWhenInvalid(t *testing.T) {
	l := Lambda{
		store:    &mockStore{get: shared.Lpa{Uid: "1"}},
		verifier: &mockVerifier{ok: true},
		logger:   logging.New(io.Discard, ""),
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"address":{"line1":"x","town":"y","country":"ZZ"}}`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	assert.JSONEq(t, `{"code":"INVALID_REQUEST","detail":"Invalid request","errors":[{"source":"/signedAt","detail":"field is required"},{"source":"/contactLanguagePreference","detail":"field is required"}]}`, resp.Body)
}

func TestHandleEventWhenRequestJsonBad(t *testing.T) {
	var buf bytes.Buffer
	l := Lambda{
		store:    &mockStore{get: shared.Lpa{Uid: "1"}},
		verifier: &mockVerifier{ok: true},
		logger:   logging.New(&buf, ""),
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.JSONEq(t, `{"code":"INTERNAL_SERVER_ERROR","detail":"Internal server error"}`, resp.Body)
	assert.Contains(t, buf.String(), `"unexpected end of JSON input"`)
}

func TestHandleEventWhenUidNotFound(t *testing.T) {
	l := Lambda{
		store:    &mockStore{},
		verifier: &mockVerifier{ok: true},
		logger:   logging.New(io.Discard, ""),
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode)
	assert.JSONEq(t, `{"code":"NOT_FOUND","detail":"Record not found"}`, resp.Body)
}

func TestHandleEventWhenStoreGetErrors(t *testing.T) {
	var buf bytes.Buffer
	l := Lambda{
		store:    &mockStore{getErr: expectedError},
		verifier: &mockVerifier{ok: true},
		logger:   logging.New(&buf, ""),
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.JSONEq(t, `{"code":"INTERNAL_SERVER_ERROR","detail":"Internal server error"}`, resp.Body)
	assert.Contains(t, buf.String(), `"expected"`)
}

func TestHandleEventWhenHeaderNotVerified(t *testing.T) {
	var buf bytes.Buffer
	l := Lambda{
		verifier: &mockVerifier{ok: false},
		logger:   logging.New(&buf, ""),
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 401, resp.StatusCode)
	assert.JSONEq(t, `{"code":"UNAUTHORISED","detail":"Invalid JWT"}`, resp.Body)
	assert.Contains(t, buf.String(), `"Unable to verify JWT from header"`)
}
