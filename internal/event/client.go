package event

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

const source = "opg.poas.lpastore"

type EventBridgeClient interface {
	PutEvents(ctx context.Context, params *eventbridge.PutEventsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error)
}

type Client struct {
	eventBusName string
	svc          EventBridgeClient
}

func NewClient(cfg aws.Config, eventBusName string) *Client {
	return &Client{
		svc:          eventbridge.NewFromConfig(cfg),
		eventBusName: eventBusName,
	}
}

func (c *Client) SendLpaUpdated(ctx context.Context, event LpaUpdated, metric *Metric) error {
	v, err := json.Marshal(event)
	if err != nil {
		return err
	}

	entries := []types.PutEventsRequestEntry{{
		EventBusName: aws.String(c.eventBusName),
		Source:       aws.String(source),
		DetailType:   aws.String("lpa-updated"),
		Detail:       aws.String(string(v)),
	}}

	if metric != nil {
		metricData, err := json.Marshal(Metrics{Metrics: []*Metric{metric}})
		if err != nil {
			return err
		}

		entries = append(entries, types.PutEventsRequestEntry{
			EventBusName: aws.String(c.eventBusName),
			Source:       aws.String(source),
			DetailType:   aws.String("metric"),
			Detail:       aws.String(string(metricData)),
		})
	}

	_, err = c.svc.PutEvents(ctx, &eventbridge.PutEventsInput{
		Entries: entries,
	})

	return err
}
