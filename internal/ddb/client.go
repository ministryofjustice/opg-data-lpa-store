package ddb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

type Client struct {
	ddb       *dynamodb.DynamoDB
	tableName string
}

func (c *Client) Put(ctx context.Context, data any) error {
	item, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		return err
	}

	_, err = c.ddb.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.tableName),
		Item:      item,
	})

	return err
}

func (c *Client) Get(ctx context.Context, uid string) (shared.Lpa, error) {
	lpa := shared.Lpa{}

	marshalledUid, err := dynamodbattribute.Marshal(uid)
	if err != nil {
		return lpa, err
	}

	getItemOutput, err := c.ddb.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(c.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"uid": marshalledUid,
		},
	})

	if err != nil {
		return lpa, err
	}

	err = dynamodbattribute.UnmarshalMap(getItemOutput.Item, &lpa)

	return lpa, err
}

func New(endpoint, tableName string) *Client {
	sess := session.Must(session.NewSession())
	sess.Config.Endpoint = &endpoint

	c := &Client{
		ddb:       dynamodb.New(sess),
		tableName: tableName,
	}

	xray.AWS(c.ddb.Client)

	return c
}
