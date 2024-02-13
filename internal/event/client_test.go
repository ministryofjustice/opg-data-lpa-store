package event

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

type mockEventBridgeClient struct {
	mock.Mock
}

func (_m *mockEventBridgeClient) PutEvents(ctx context.Context, params *eventbridge.PutEventsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error) {
	var r0 *eventbridge.PutEventsOutput
	var r1 error = errors.New("err")

	return r0, r1
}

func TestClientSendEvent(t *testing.T) {
	ctx := context.Background()
	expectedError := errors.New("err")

	testcases := map[string]func() (func(*Client) error, any) {
		"lpa-updated": func() (func(*Client) error, any) {
			event := LpaUpdated{ uid: "M-1234-1234-1234", changeType: "CREATED" }

			return func(client *Client) error { return client.SendLpaUpdated(ctx, event) }, event
		},
	}

	for eventName, setup := range testcases {
		t.Run(eventName, func(t *testing.T) {
			fn, event := setup()
			data, _ := json.Marshal(event)

			mockClient := &mockEventBridgeClient{}
			mockClient.On("PutEvents", mock.Anything, &eventbridge.PutEventsInput{
					Entries: []types.PutEventsRequestEntry{{
						EventBusName: aws.String("my-bus"),
						Source:       aws.String("opg.poas.lpastore"),
						DetailType:   aws.String(eventName),
						Detail:       aws.String(string(data)),
					}},
				}).
				Return(nil, expectedError)

			svc := &Client{svc: mockClient, eventBusName: "my-bus"}
			err := fn(svc)

			assert.Equal(t, expectedError, err)
		})
	}
}
