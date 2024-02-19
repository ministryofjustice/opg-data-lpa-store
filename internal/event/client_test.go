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
	"github.com/stretchr/testify/mock"
)

type mockEventbridgeClient struct {
	mock.Mock
}

func (_m *mockEventbridgeClient) PutEvents(ctx context.Context, params *eventbridge.PutEventsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error) {
	var r0 *eventbridge.PutEventsOutput
	var r1 error = errors.New("err")

	return r0, r1
}

func TestClientSendEvent(t *testing.T) {
	ctx := context.Background()
	expectedError := errors.New("err")

	event := LpaUpdated{ Uid: "M-1234-1234-1234", ChangeType: "CREATED" }

	fn, event := func(client *Client) error { return client.SendLpaUpdated(ctx, event) }, event

	data, _ := json.Marshal(event)

	mockClient := &mockEventbridgeClient{}
	mockClient.On("PutEvents", mock.Anything, &eventbridge.PutEventsInput{
			Entries: []types.PutEventsRequestEntry{{
				EventBusName: aws.String("my-bus"),
				Source:       aws.String("opg.poas.lpastore"),
				DetailType:   aws.String("lpa-updated"),
				Detail:       aws.String(string(data)),
			}},
		}).
		Return(nil, expectedError)

	svc := &Client{svc: mockClient, eventBusName: "my-bus"}
	err := fn(svc)

	assert.Equal(t, expectedError, err)
}
