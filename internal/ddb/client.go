package ddb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

type Client struct {
	ddb              *dynamodb.Client
	tableName        string
	changesTableName string
}

func New(cfg aws.Config, tableName, changesTableName string) *Client {
	return &Client{
		ddb:              dynamodb.NewFromConfig(cfg),
		tableName:        tableName,
		changesTableName: changesTableName,
	}
}

func (c *Client) PutChanges(ctx context.Context, data any, update shared.Update) error {
	changesItem, _ := attributevalue.MarshalMap(map[string]interface{}{
		"uid":     update.Uid,
		"applied": update.Applied,
		"author":  update.Author,
		"type":    update.Type,
		"change":  update.Changes,
	})

	item, err := attributevalue.MarshalMapWithOptions(data, encoderOptions)
	if err != nil {
		return err
	}

	transactInput := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			// write the LPA to the deeds table
			{
				Put: &types.Put{
					TableName: aws.String(c.tableName),
					Item:      item,
				},
			},

			// record the change
			{
				Put: &types.Put{
					TableName: aws.String(c.changesTableName),
					Item:      changesItem,
				},
			},
		},
	}

	_, err = c.ddb.TransactWriteItems(ctx, transactInput)

	return err
}

func (c *Client) Put(ctx context.Context, data any) error {
	item, err := attributevalue.MarshalMapWithOptions(data, encoderOptions)
	if err != nil {
		return err
	}

	_, err = c.ddb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.tableName),
		Item:      item,
	})

	return err
}

func (c *Client) Get(ctx context.Context, uid string) (shared.Lpa, error) {
	lpa := shared.Lpa{}

	marshalledUid, err := attributevalue.Marshal(uid)
	if err != nil {
		return lpa, err
	}

	getItemOutput, err := c.ddb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(c.tableName),
		Key: map[string]types.AttributeValue{
			"uid": marshalledUid,
		},
	})

	if err != nil {
		return lpa, err
	}

	err = attributevalue.UnmarshalMapWithOptions(getItemOutput.Item, &lpa, decoderOptions)

	return lpa, err
}

func decoderOptions(opts *attributevalue.DecoderOptions) {
	opts.TagKey = "json"
}

func encoderOptions(opts *attributevalue.EncoderOptions) {
	opts.TagKey = "json"
}
