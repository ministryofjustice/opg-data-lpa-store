package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/event"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var (
	ctx           = context.WithValue(context.Background(), (*string)(nil), "testing")
	expectedError = errors.New("err")
	testNow       = time.Date(2024, time.January, 2, 12, 13, 14, 15, time.UTC)
	testNowFn     = func() time.Time { return testNow }
	validLpaInit  = shared.LpaInit{
		LpaType: shared.LpaTypePropertyAndAffairs,
		Channel: shared.ChannelOnline,
		Donor: shared.Donor{
			Person: shared.Person{
				UID:        "a06daa09-750d-4e02-9877-0ea782491014",
				FirstNames: "donor-firstname",
				LastName:   "donor-lastname",
			},
			Address: shared.Address{
				Line1:   "donor-line1",
				Country: "GB",
			},
			DateOfBirth:               makeDate("2020-01-02"),
			Email:                     "donor-email",
			ContactLanguagePreference: shared.LangEn,
		},
		Attorneys: []shared.Attorney{{
			Person: shared.Person{
				UID:        "c442a9a2-9d14-48ca-9cfa-d30d591b0d68",
				FirstNames: "attorney-firstname",
				LastName:   "attorney-lastname",
			},
			Address: shared.Address{
				Line1:   "attorney-line1",
				Country: "GB",
			},
			AppointmentType:           shared.AppointmentTypeOriginal,
			DateOfBirth:               makeDate("2020-02-03"),
			ContactLanguagePreference: shared.LangEn,
			Status:                    shared.AttorneyStatusActive,
			Channel:                   shared.ChannelPaper,
		}},
		CertificateProvider: shared.CertificateProvider{
			Person: shared.Person{
				UID:        "e9751c0a-0504-4ec6-942e-b85fddbbd178",
				FirstNames: "certificate-provider-firstname",
				LastName:   "certificate-provider-lastname",
			},
			Address: shared.Address{
				Line1:   "certificate-provider-line1",
				Country: "GB",
			},
			Channel:                   shared.ChannelOnline,
			Email:                     "certificate-provider-email",
			Phone:                     "0777777777",
			ContactLanguagePreference: shared.LangEn,
		},
		WhenTheLpaCanBeUsed:              shared.CanUseWhenHasCapacity,
		SignedAt:                         testNow,
		WitnessedByCertificateProviderAt: testNow,
	}
)

func makeDate(s string) shared.Date {
	d := &shared.Date{}
	_ = d.UnmarshalText([]byte(s))
	return *d
}

func TestLambdaHandleEvent(t *testing.T) {
	for _, channel := range []shared.Channel{shared.ChannelOnline, shared.ChannelPaper} {
		t.Run(string(channel), func(t *testing.T) {
			lpaInit := validLpaInit
			lpaInit.Channel = channel
			body, _ := json.Marshal(lpaInit)

			lpa := shared.Lpa{
				Uid:       "my-uid",
				Status:    shared.LpaStatusInProgress,
				UpdatedAt: testNow,
				LpaInit:   lpaInit,
			}

			req := events.APIGatewayProxyRequest{
				PathParameters: map[string]string{"uid": "my-uid"},
				Body:           string(body),
			}

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
				Return(shared.Lpa{}, nil)
			store.EXPECT().
				Put(ctx, lpa).
				Return(nil)

			staticLpaStorage := newMockS3Client(t)
			staticLpaStorage.EXPECT().
				Put(ctx, "my-uid/donor-executed-lpa.json", lpa).
				Return(nil)

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendLpaUpdated(ctx, event.LpaUpdated{
					Uid:        "my-uid",
					ChangeType: "CREATE",
				}).
				Return(nil)

			lambda := &Lambda{
				verifier:         verifier,
				logger:           logger,
				store:            store,
				staticLpaStorage: staticLpaStorage,
				eventClient:      eventClient,
				now:              testNowFn,
			}

			resp, err := lambda.HandleEvent(ctx, req)
			assert.Nil(t, err)
			assert.Equal(t, events.APIGatewayProxyResponse{
				StatusCode: 201,
				Body:       "{}",
			}, resp)
		})
	}
}

func TestLambdaHandleEventWhenPaperSubmissionContainsImages(t *testing.T) {
	lpaInit := validLpaInit
	lpaInit.Channel = shared.ChannelPaper
	lpaInit.RestrictionsAndConditionsImages = []shared.FileUpload{{
		Filename: "restriction.jpg",
		Data:     "some-base64",
	}}
	body, _ := json.Marshal(lpaInit)

	lpa := shared.Lpa{
		Uid:       "my-uid",
		Status:    shared.LpaStatusInProgress,
		UpdatedAt: testNow,
		LpaInit:   lpaInit,
	}
	lpa.RestrictionsAndConditionsImages = []shared.File{{Path: "a", Hash: "b"}}

	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"uid": "my-uid"},
		Body:           string(body),
	}

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
		Return(shared.Lpa{}, nil)
	store.EXPECT().
		Put(ctx, lpa).
		Return(nil)

	staticLpaStorage := newMockS3Client(t)
	staticLpaStorage.EXPECT().
		Put(ctx, "my-uid/donor-executed-lpa.json", lpa).
		Return(nil)
	staticLpaStorage.EXPECT().
		UploadFile(ctx, shared.FileUpload{Filename: "restriction.jpg", Data: "some-base64"}, "my-uid/scans/rc_0_restriction.jpg").
		Return(shared.File{Path: "a", Hash: "b"}, nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendLpaUpdated(ctx, event.LpaUpdated{
			Uid:        "my-uid",
			ChangeType: "CREATE",
		}).
		Return(nil)

	lambda := &Lambda{
		verifier:         verifier,
		logger:           logger,
		store:            store,
		staticLpaStorage: staticLpaStorage,
		eventClient:      eventClient,
		now:              testNowFn,
	}

	resp, err := lambda.HandleEvent(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: 201,
		Body:       "{}",
	}, resp)
}

func TestLambdaHandleEventWhenPaperSubmissionHasValidationErrors(t *testing.T) {
	lpaInit := validLpaInit
	lpaInit.Channel = shared.ChannelPaper
	lpaInit.WhenTheLpaCanBeUsed = shared.CanUseUnset
	body, _ := json.Marshal(lpaInit)

	lpa := shared.Lpa{
		Uid:       "my-uid",
		Status:    shared.LpaStatusInProgress,
		UpdatedAt: testNow,
		LpaInit:   lpaInit,
	}

	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"uid": "my-uid"},
		Body:           string(body),
	}

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(req).
		Return(nil, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header")
	logger.EXPECT().
		Info("encountered validation errors in lpa", slog.String("uid", "my-uid"))

	store := newMockStore(t)
	store.EXPECT().
		Get(ctx, "my-uid").
		Return(shared.Lpa{}, nil)
	store.EXPECT().
		Put(ctx, lpa).
		Return(nil)

	staticLpaStorage := newMockS3Client(t)
	staticLpaStorage.EXPECT().
		Put(ctx, "my-uid/donor-executed-lpa.json", lpa).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendLpaUpdated(ctx, event.LpaUpdated{
			Uid:        "my-uid",
			ChangeType: "CREATE",
		}).
		Return(nil)

	lambda := &Lambda{
		verifier:         verifier,
		logger:           logger,
		store:            store,
		staticLpaStorage: staticLpaStorage,
		eventClient:      eventClient,
		now:              testNowFn,
	}

	resp, err := lambda.HandleEvent(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: 201,
		Body:       "{}",
	}, resp)
}

func TestLambdaHandleEventWhenUnauthorised(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"uid": "my-uid"},
		Body:           "{}",
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
		now:      testNowFn,
	}

	resp, err := lambda.HandleEvent(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: 401,
		Body:       `{"code":"UNAUTHORISED","detail":"Invalid JWT"}`,
	}, resp)
}

func TestLambdaHandleEventWhenLpaAlreadyExists(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"uid": "my-uid"},
		Body:           "{}",
	}

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
		Return(shared.Lpa{Uid: "my-uid"}, nil)

	lambda := &Lambda{
		verifier: verifier,
		logger:   logger,
		store:    store,
		now:      testNowFn,
	}

	resp, err := lambda.HandleEvent(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: 400,
		Body:       `{"code":"INVALID_REQUEST","detail":"LPA with UID already exists"}`,
	}, resp)
}

func TestLambdaHandleEventWhenUploadFileErrors(t *testing.T) {
	lpaInit := validLpaInit
	lpaInit.Channel = shared.ChannelPaper
	lpaInit.RestrictionsAndConditionsImages = []shared.FileUpload{{
		Filename: "restriction.jpg",
		Data:     "some-base64",
	}}
	body, _ := json.Marshal(lpaInit)

	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"uid": "my-uid"},
		Body:           string(body),
	}

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(req).
		Return(nil, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header")
	logger.EXPECT().
		Error("error saving restrictions and conditions image", slog.Any("err", expectedError))

	store := newMockStore(t)
	store.EXPECT().
		Get(ctx, "my-uid").
		Return(shared.Lpa{}, nil)

	staticLpaStorage := newMockS3Client(t)
	staticLpaStorage.EXPECT().
		UploadFile(ctx, shared.FileUpload{Filename: "restriction.jpg", Data: "some-base64"}, "my-uid/scans/rc_0_restriction.jpg").
		Return(shared.File{}, expectedError)

	lambda := &Lambda{
		verifier:         verifier,
		logger:           logger,
		store:            store,
		staticLpaStorage: staticLpaStorage,
		now:              testNowFn,
	}

	resp, err := lambda.HandleEvent(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       `{"code":"INTERNAL_SERVER_ERROR","detail":"Internal server error"}`,
	}, resp)
}

func TestLambdaHandleEventWhenSendLpaUpdatedErrors(t *testing.T) {
	lpaInit := validLpaInit
	body, _ := json.Marshal(lpaInit)

	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"uid": "my-uid"},
		Body:           string(body),
	}

	verifier := newMockVerifier(t)
	verifier.EXPECT().
		VerifyHeader(req).
		Return(nil, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Debug("Successfully parsed JWT from event header")
	logger.EXPECT().
		Error("unexpected error occurred", slog.Any("err", expectedError))

	store := newMockStore(t)
	store.EXPECT().
		Get(ctx, "my-uid").
		Return(shared.Lpa{}, nil)
	store.EXPECT().
		Put(ctx, mock.Anything).
		Return(nil)

	staticLpaStorage := newMockS3Client(t)
	staticLpaStorage.EXPECT().
		Put(ctx, "my-uid/donor-executed-lpa.json", mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendLpaUpdated(ctx, event.LpaUpdated{
			Uid:        "my-uid",
			ChangeType: "CREATE",
		}).
		Return(expectedError)

	lambda := &Lambda{
		verifier:         verifier,
		logger:           logger,
		store:            store,
		staticLpaStorage: staticLpaStorage,
		eventClient:      eventClient,
		now:              testNowFn,
	}

	resp, err := lambda.HandleEvent(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: 201,
		Body:       "{}",
	}, resp)
}
