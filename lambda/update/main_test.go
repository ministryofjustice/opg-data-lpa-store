package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/event"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	errExpected = errors.New("expected")
	jsonNull    = json.RawMessage("null")

	testNow   = time.Date(2024, time.January, 2, 12, 13, 14, 15, time.UTC)
	testNowFn = func() time.Time { return testNow }
)

func newAllowedMockVerifier(t *testing.T) *mockVerifier {
	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(mock.Anything).
		Return(&shared.LpaStoreClaims{}, nil)
	return verifier
}

func TestHandleEvent(t *testing.T) {
	signedAt := time.Date(2022, time.January, 2, 12, 13, 14, 6, time.UTC)

	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header", mock.Anything)

	store := newMockStore(t)
	store.EXPECT().
		Get(mock.Anything, mock.Anything).
		Return(shared.Lpa{
			Uid: "1",
			LpaInit: shared.LpaInit{
				CertificateProvider: shared.CertificateProvider{
					Email:   "a@example.com",
					Channel: shared.ChannelPaper,
				},
			},
		}, nil)
	store.EXPECT().
		PutChanges(mock.Anything, shared.Lpa{
			Uid: "1",
			LpaInit: shared.LpaInit{
				CertificateProvider: shared.CertificateProvider{
					SignedAt:                  &signedAt,
					ContactLanguagePreference: shared.LangEn,
					Channel:                   shared.ChannelOnline,
					Email:                     "b@example.com",
				},
			},
		}, mock.MatchedBy(func(update shared.Update) bool {
			id := update.Id
			applied := update.Applied
			update.Id = ""
			update.Applied = ""

			return assert.NoError(t, uuid.Validate(id)) &&
				assert.Regexp(t, regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`), applied) &&
				assert.Equal(t, shared.Update{
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
				}, update)
		})).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendLpaUpdated(mock.Anything, event.LpaUpdated{
			Uid:        "1",
			ChangeType: "CERTIFICATE_PROVIDER_SIGN",
		}, &event.Metric{
			Project:          "MRLPA",
			Category:         "metric",
			Subcategory:      "FunnelCompletionRate",
			Environment:      "ENVIRONMENT",
			MeasureName:      "CERTIFICATEPROVIDER",
			MeasureValue:     "1",
			MeasureValueType: "BIGINT",
			Time:             strconv.FormatInt(testNow.UnixMilli(), 10),
		}).
		Return(nil)

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(mock.Anything).
		Return(&shared.LpaStoreClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "1234",
			},
		}, nil)

	l := Lambda{
		eventClient: eventClient,
		store:       store,
		verifier:    verifier,
		environment: "ENVIRONMENT",
		logger:      logger,
		now:         testNowFn,
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"type":"CERTIFICATE_PROVIDER_SIGN","changes":[{"key":"/certificateProvider/signedAt","old":null,"new":"2022-01-02T12:13:14.000000006Z"},{"key":"/certificateProvider/contactLanguagePreference","old":null,"new":"en"},{"key":"/certificateProvider/email","old":"a@example.com","new":"b@example.com"},{"key":"/certificateProvider/channel","old":"paper","new":"online"}]}`,
	})

	assert.Nil(t, err)
	assert.Equal(t, 201, resp.StatusCode)
	assert.Contains(t, resp.Body, `"2022-01-02T12:13:14.000000006Z"`)
	assert.Contains(t, resp.Body, `"en"`)
}

func TestHandleEventWhenUnknownType(t *testing.T) {
	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header", mock.Anything)

	store := newMockStore(t)
	store.EXPECT().
		Get(mock.Anything, mock.Anything).
		Return(shared.Lpa{Uid: "1"}, nil)

	l := Lambda{
		store:       store,
		verifier:    newAllowedMockVerifier(t),
		logger:      logger,
		eventClient: newMockEventClient(t),
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"type":"SCANNING_CORRECTION","changes":[{"key":"/donor/firstNames","old":"Johm","new":"John"}]}`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	assert.JSONEq(t, `{"code":"INVALID_REQUEST","detail":"Invalid request","errors":[{"source":"/type","detail":"invalid value"}]}`, resp.Body)
}

func TestHandleEventWhenUpdateInvalid(t *testing.T) {
	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header", mock.Anything)

	store := newMockStore(t)
	store.EXPECT().
		Get(mock.Anything, mock.Anything).
		Return(shared.Lpa{Uid: "1"}, nil)

	l := Lambda{
		store:    store,
		verifier: newAllowedMockVerifier(t),
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
	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header", mock.Anything)
	logger.EXPECT().
		Debug("Uid not found", mock.Anything)

	store := newMockStore(t)
	store.EXPECT().
		Get(mock.Anything, mock.Anything).
		Return(shared.Lpa{}, nil)

	l := Lambda{
		store:    store,
		verifier: newAllowedMockVerifier(t),
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
	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header", mock.Anything)
	logger.EXPECT().
		Error("error fetching LPA", slog.Any("err", errExpected))

	store := newMockStore(t)
	store.EXPECT().
		Get(mock.Anything, mock.Anything).
		Return(shared.Lpa{}, errExpected)

	l := Lambda{
		store:    store,
		verifier: newAllowedMockVerifier(t),
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
	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header", mock.Anything)
	logger.EXPECT().
		Error("error unmarshalling request", mock.Anything)

	l := Lambda{
		verifier: newAllowedMockVerifier(t),
		logger:   logger,
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.JSONEq(t, `{"code":"INTERNAL_SERVER_ERROR","detail":"Internal server error"}`, resp.Body)
}

func TestHandleEventWhenHeaderNotVerified(t *testing.T) {
	logger := newMockLogger(t)
	logger.EXPECT().
		Info("Unable to verify JWT from header", mock.Anything)

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(mock.Anything).
		Return(nil, errors.New("Invalid JWT"))

	l := Lambda{
		verifier: verifier,
		logger:   logger,
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 401, resp.StatusCode)
	assert.JSONEq(t, `{"code":"UNAUTHORISED","detail":"Invalid JWT"}`, resp.Body)
}

func TestHandleEventWhenSendLpaUpdatedFailed(t *testing.T) {
	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header", mock.Anything)
	logger.EXPECT().
		Error("unexpected error occurred", slog.Any("err", errors.New("Update failed")))

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(mock.Anything).
		Return(&shared.LpaStoreClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "1234",
			},
		}, nil)

	store := newMockStore(t)
	store.EXPECT().
		Get(mock.Anything, mock.Anything).
		Return(shared.Lpa{Uid: "1"}, nil)
	store.EXPECT().
		PutChanges(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	client := newMockEventClient(t)
	client.EXPECT().
		SendLpaUpdated(mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("Update failed"))

	l := Lambda{
		eventClient: client,
		store:       store,
		verifier:    verifier,
		logger:      logger,
		now:         testNowFn,
	}

	resp, err := l.HandleEvent(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"type":"CERTIFICATE_PROVIDER_SIGN","changes":[{"key":"/certificateProvider/signedAt","old":null,"new":"2022-01-02T12:13:14.000000006Z"},{"key":"/certificateProvider/contactLanguagePreference","old":null,"new":"en"},{"key":"/certificateProvider/email","old":null,"new":"a@example.com"}]}`,
	})

	assert.Nil(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}
