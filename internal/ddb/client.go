package ddb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

type dynamodbClient interface {
	TransactWriteItems(ctx context.Context, params *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	BatchGetItem(ctx context.Context, params *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

type QueryPaginator interface {
	HasMorePages() bool
	NextPage(ctx context.Context) (*dynamodb.QueryOutput, error)
}

type PaginatorFactory interface {
	NewQueryPaginator(input *dynamodb.QueryInput) QueryPaginator
}

type awsPaginator struct {
	p *dynamodb.QueryPaginator
}

func (a *awsPaginator) HasMorePages() bool {
	return a.p.HasMorePages()
}

func (a *awsPaginator) NextPage(ctx context.Context) (*dynamodb.QueryOutput, error) {
	return a.p.NextPage(ctx)
}

type awsPaginatorFactory struct {
	svc *dynamodb.Client
}

func (f *awsPaginatorFactory) NewQueryPaginator(input *dynamodb.QueryInput) QueryPaginator {
	return &awsPaginator{
		p: dynamodb.NewQueryPaginator(f.svc, input),
	}
}

type Client struct {
	svc              dynamodbClient
	tableName        string
	changesTableName string
	paginatorFactory PaginatorFactory
}

func New(cfg aws.Config, tableName, changesTableName string) *Client {
	svc := dynamodb.NewFromConfig(cfg)

	return &Client{
		svc:              svc,
		tableName:        tableName,
		changesTableName: changesTableName,
		paginatorFactory: &awsPaginatorFactory{svc: svc},
	}
}

func (c *Client) PutChanges(ctx context.Context, data any, update shared.Update) error {
	changesItem, _ := attributevalue.MarshalMap(map[string]interface{}{
		"uid":     update.Uid,
		"applied": update.Applied,
		"author":  update.Author,
		"type":    update.Type,
		"changes": update.Changes,
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

	_, err = c.svc.TransactWriteItems(ctx, transactInput)

	return err
}

func (c *Client) Put(ctx context.Context, data any) error {
	item, err := attributevalue.MarshalMapWithOptions(data, encoderOptions)
	if err != nil {
		return err
	}

	_, err = c.svc.PutItem(ctx, &dynamodb.PutItemInput{
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

	getItemOutput, err := c.svc.GetItem(ctx, &dynamodb.GetItemInput{
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

func (c *Client) GetChanges(ctx context.Context, uid string) ([]shared.Update, error) {
	var response *dynamodb.QueryOutput
	var updates []shared.Update

	keyEx := expression.Key("uid").Equal(expression.Value(uid))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()

	if err != nil {
		return nil, err
	}

	scanIndexForward := false
	queryPaginator := c.paginatorFactory.NewQueryPaginator(&dynamodb.QueryInput{
		TableName:                 aws.String(c.changesTableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ScanIndexForward:          &scanIndexForward,
	})
	for queryPaginator.HasMorePages() {
		response, err = queryPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		} else {
			var updatesPage []shared.Update
			err = attributevalue.UnmarshalListOfMaps(response.Items, &updatesPage)
			if err != nil {
				return nil, err
			} else {
				updates = append(updates, updatesPage...)
			}
		}
	}

	return updates, nil
}

func (c *Client) GetList(ctx context.Context, uids []string) ([]shared.Lpa, error) {
	keys := make([]map[string]types.AttributeValue, len(uids))
	for i, uid := range uids {
		keys[i] = map[string]types.AttributeValue{
			"uid": &types.AttributeValueMemberS{Value: uid},
		}
	}

	output, err := c.svc.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			c.tableName: {
				Keys: keys,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var v []shared.Lpa
	if err := attributevalue.UnmarshalListOfMaps(output.Responses[c.tableName], &v); err != nil {
		return nil, err
	}

	return v, nil
}

func decoderOptions(opts *attributevalue.DecoderOptions) {
	opts.TagKey = "json"
}

func encoderOptions(opts *attributevalue.EncoderOptions) {
	opts.TagKey = "json"
}
