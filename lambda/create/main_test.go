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
	"github.com/ministryofjustice/opg-data-lpa-store/internal/event"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	ctx          = context.WithValue(context.Background(), (*string)(nil), "testing")
	errExample   = errors.New("err")
	testNow      = time.Date(2024, time.January, 2, 12, 13, 14, 15, time.UTC)
	testNowFn    = func() time.Time { return testNow }
	validLpaInit = shared.LpaInit{
		LpaType:  shared.LpaTypePropertyAndAffairs,
		Channel:  shared.ChannelOnline,
		Language: shared.LangEn,
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
	onlineWithDefault := validLpaInit
	onlineWithDefault.WhenTheLpaCanBeUsed = shared.CanUseUnset

	onlineWithDefaultLpa := validLpaInit
	onlineWithDefaultLpa.WhenTheLpaCanBeUsed = shared.CanUseWhenHasCapacity
	onlineWithDefaultLpa.WhenTheLpaCanBeUsedIsDefault = true

	paperLpaInit := validLpaInit
	paperLpaInit.Channel = shared.ChannelPaper

	paperWithDefault := paperLpaInit
	paperWithDefault.WhenTheLpaCanBeUsed = shared.CanUseUnset

	paperWithDefaultLpa := paperLpaInit
	paperWithDefaultLpa.WhenTheLpaCanBeUsed = shared.CanUseWhenHasCapacity
	paperWithDefaultLpa.WhenTheLpaCanBeUsedIsDefault = true

	testcases := map[string]struct {
		input       shared.LpaInit
		measureName string
		lpa         shared.Lpa
	}{
		"online": {
			input:       validLpaInit,
			measureName: "ONLINEDONOR",
			lpa: shared.Lpa{
				Uid:       "my-uid",
				Status:    shared.LpaStatusInProgress,
				UpdatedAt: testNow,
				LpaInit:   validLpaInit,
			},
		},
		"online with default": {
			input:       onlineWithDefault,
			measureName: "ONLINEDONOR",
			lpa: shared.Lpa{
				Uid:       "my-uid",
				Status:    shared.LpaStatusInProgress,
				UpdatedAt: testNow,
				LpaInit:   onlineWithDefaultLpa,
			},
		},
		"paper": {
			input:       paperLpaInit,
			measureName: "PAPERDONOR",
			lpa: shared.Lpa{
				Uid:       "my-uid",
				Status:    shared.LpaStatusInProgress,
				UpdatedAt: testNow,
				LpaInit:   paperLpaInit,
			},
		},
		"paper with default": {
			input:       paperWithDefault,
			measureName: "PAPERDONOR",
			lpa: shared.Lpa{
				Uid:       "my-uid",
				Status:    shared.LpaStatusInProgress,
				UpdatedAt: testNow,
				LpaInit:   paperWithDefaultLpa,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			body, _ := json.Marshal(tc.input)

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
				Put(ctx, tc.lpa).
				Return(nil)

			staticLpaStorage := newMockS3Client(t)
			staticLpaStorage.EXPECT().
				Put(ctx, "my-uid/donor-executed-lpa.json", tc.lpa).
				Return(nil)

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendLpaUpdated(ctx, event.LpaUpdated{
					Uid:        "my-uid",
					ChangeType: "CREATE",
				}, &event.Metric{
					Project:          "MRLPA",
					Category:         "metric",
					Subcategory:      "FunnelCompletionRate",
					Environment:      "E",
					MeasureName:      tc.measureName,
					MeasureValue:     "1",
					MeasureValueType: "BIGINT",
					Time:             strconv.FormatInt(testNow.UnixMilli(), 10),
				}).
				Return(nil)

			lambda := &Lambda{
				verifier:         verifier,
				logger:           logger,
				store:            store,
				staticLpaStorage: staticLpaStorage,
				eventClient:      eventClient,
				now:              testNowFn,
				environment:      "E",
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
		}, mock.Anything).
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
	lpaInit.WhenTheLpaCanBeUsed = shared.CanUse("bad")
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
		}, mock.Anything).
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
		Return(nil, errExample)

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
		Error("error saving restrictions and conditions image", slog.Any("err", errExample))

	store := newMockStore(t)
	store.EXPECT().
		Get(ctx, "my-uid").
		Return(shared.Lpa{}, nil)

	staticLpaStorage := newMockS3Client(t)
	staticLpaStorage.EXPECT().
		UploadFile(ctx, shared.FileUpload{Filename: "restriction.jpg", Data: "some-base64"}, "my-uid/scans/rc_0_restriction.jpg").
		Return(shared.File{}, errExample)

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
		Error("unexpected error occurred", slog.Any("err", errExample))

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
		}, mock.Anything).
		Return(errExample)

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

func TestLambdaHandleEventAddsActorUids(t *testing.T) {
	body, _ := json.Marshal(shared.LpaInit{
		Channel: shared.ChannelPaper,
		Donor: shared.Donor{
			Person: shared.Person{
				FirstNames: "donor-firstname",
				LastName:   "donor-lastname",
			},
		},
		CertificateProvider: shared.CertificateProvider{
			Person: shared.Person{
				FirstNames: "certificate-provider-firstname",
				LastName:   "certificate-provider-lastname",
			},
		},
		Attorneys: []shared.Attorney{
			{
				Person: shared.Person{
					FirstNames: "attorney-firstname",
					LastName:   "attorney-lastname",
				},
				AppointmentType: shared.AppointmentTypeOriginal,
			},
			{
				Person: shared.Person{
					FirstNames: "attorney2-firstname",
					LastName:   "attorney2-lastname",
				},
				AppointmentType: shared.AppointmentTypeReplacement,
			},
		},
		TrustCorporations: []shared.TrustCorporation{
			{},
			{UID: "76fca433-639c-4119-8b3b-6f1e5d82de55"},
		},
		PeopleToNotify: []shared.PersonToNotify{
			{Person: shared.Person{UID: "752f2031-66c6-412d-b8b8-3f9d56bb6e86"}},
			{},
		},
		IndependentWitness:  &shared.IndependentWitness{},
		AuthorisedSignatory: &shared.AuthorisedSignatory{},
	})

	uuidRegex := regexp.MustCompile("^[0-f]{8}-[0-f]{4}-[0-f]{4}-[0-f]{4}-[0-f]{12}$")

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
		Info("encountered validation errors in lpa", mock.Anything)

	store := newMockStore(t)
	store.EXPECT().
		Get(ctx, "my-uid").
		Return(shared.Lpa{}, nil)
	store.EXPECT().
		Put(ctx, mock.MatchedBy(func(lpa shared.Lpa) bool {
			return uuidRegex.MatchString(lpa.Donor.UID) &&
				uuidRegex.MatchString(lpa.CertificateProvider.UID) &&
				uuidRegex.MatchString(lpa.Attorneys[0].UID) &&
				uuidRegex.MatchString(lpa.Attorneys[1].UID) &&
				uuidRegex.MatchString(lpa.TrustCorporations[0].UID) &&
				lpa.TrustCorporations[1].UID == "76fca433-639c-4119-8b3b-6f1e5d82de55" &&
				lpa.PeopleToNotify[0].UID == "752f2031-66c6-412d-b8b8-3f9d56bb6e86" &&
				uuidRegex.MatchString(lpa.PeopleToNotify[1].UID) &&
				uuidRegex.MatchString(lpa.IndependentWitness.UID) &&
				uuidRegex.MatchString(lpa.AuthorisedSignatory.UID)
		})).
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
		}, &event.Metric{
			Project:          "MRLPA",
			Category:         "metric",
			Subcategory:      "FunnelCompletionRate",
			Environment:      "E",
			MeasureName:      "PAPERDONOR",
			MeasureValue:     "1",
			MeasureValueType: "BIGINT",
			Time:             strconv.FormatInt(testNow.UnixMilli(), 10),
		}).
		Return(nil)

	lambda := &Lambda{
		verifier:         verifier,
		logger:           logger,
		store:            store,
		staticLpaStorage: staticLpaStorage,
		eventClient:      eventClient,
		environment:      "E",
		now:              testNowFn,
	}

	resp, err := lambda.HandleEvent(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: 201,
		Body:       "{}",
	}, resp)
}
