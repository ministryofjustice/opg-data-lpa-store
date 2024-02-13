package event

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

const source = "opg.poas.lpastore"

type eventBridgeClient interface {
	PutEvents(ctx context.Context, params *eventbridge.PutEventsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error)
}

type Client struct {
	eventBusName string
	svc          eventBridgeClient
}

func NewClient(cfg aws.Config, eventBusName string) *Client {
	return &Client{
		svc: eventbridge.NewFromConfig(cfg),
		eventBusName: eventBusName,
	}
}

func (c *Client) SendLpaUpdated(ctx context.Context, event LpaUpdated) error {
	return c.send(ctx, "lpa-updated", event)
}

func (c *Client) send(ctx context.Context, eventType string, detail any) error {

	v, err := json.Marshal(detail)
	if err != nil {
		return err
	}
	
	_, err = c.svc.PutEvents(ctx, &eventbridge.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{{
			EventBusName: aws.String(c.eventBusName),
			Source: aws.String(source),
			DetailType: aws.String(eventType),
			Detail: aws.String(string(v)),
		}},
	})

	return err
}
