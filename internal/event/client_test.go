package event

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
)

var (
	ctx           = context.WithValue(context.Background(), "for", "testing")
	expectedError = errors.New("err")
	eventBusName  = "an-event-bus-name"
)

func TestNewClient(t *testing.T) {
	client := NewClient(aws.Config{}, eventBusName)

	assert.IsType(t, (*eventbridge.Client)(nil), client.svc)
	assert.Equal(t, eventBusName, client.eventBusName)
}

func TestClientSendLpaUpdated(t *testing.T) {
	event := LpaUpdated{Uid: "M-1234-1234-1234", ChangeType: "CREATE"}

	eventBridgeClient := newMockEventBridgeClient(t)
	eventBridgeClient.EXPECT().
		PutEvents(ctx, &eventbridge.PutEventsInput{
			Entries: []types.PutEventsRequestEntry{{
				EventBusName: aws.String(eventBusName),
				Source:       aws.String(source),
				DetailType:   aws.String("lpa-updated"),
				Detail:       aws.String(`{"uid":"M-1234-1234-1234","changeType":"CREATE"}`),
			}},
		}).
		Return(nil, expectedError)

	client := &Client{svc: eventBridgeClient, eventBusName: eventBusName}

	err := client.SendLpaUpdated(ctx, event)
	assert.Equal(t, expectedError, err)
}
