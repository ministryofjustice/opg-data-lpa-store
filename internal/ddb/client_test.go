package ddb

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

type ctxValueType string

const ctxValue ctxValueType = "for"

var (
	ctx              = context.WithValue(context.Background(), ctxValue, "testing")
	tableName        = "a-table"
	changesTableName = "a-change-table"
	errExpected      = errors.New("hey")
)

func TestNew(t *testing.T) {
	client := New(aws.Config{}, tableName, changesTableName)

	assert.IsType(t, (*dynamodb.Client)(nil), client.svc)
	assert.Equal(t, tableName, client.tableName)
	assert.Equal(t, changesTableName, client.changesTableName)
}

func TestClientPutChanges(t *testing.T) {
	dynamodbClient := newMockDynamodbClient(t)
	dynamodbClient.EXPECT().
		TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
			TransactItems: []types.TransactWriteItem{{
				Put: &types.Put{
					TableName: aws.String(tableName),
					Item: map[string]types.AttributeValue{
						"hey": &types.AttributeValueMemberS{Value: "hello"},
					},
				},
			}, {
				Put: &types.Put{
					TableName: aws.String(changesTableName),
					Item: map[string]types.AttributeValue{
						"id":      &types.AttributeValueMemberS{Value: "123"},
						"uid":     &types.AttributeValueMemberS{Value: "a-uid"},
						"applied": &types.AttributeValueMemberS{Value: "2024-01-01Tsomething"},
						"author":  &types.AttributeValueMemberS{Value: "an-author"},
						"type":    &types.AttributeValueMemberS{Value: "a-type"},
						"changes": &types.AttributeValueMemberL{Value: []types.AttributeValue{
							&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
								"Key": &types.AttributeValueMemberS{Value: "a-key"},
								"Old": &types.AttributeValueMemberB{Value: []byte("old")},
								"New": &types.AttributeValueMemberB{Value: []byte("new")},
							}},
						}},
					},
				},
			}},
		}).
		Return(nil, errExpected)

	client := &Client{
		svc:              dynamodbClient,
		tableName:        tableName,
		changesTableName: changesTableName,
	}

	err := client.PutChanges(ctx, map[string]string{"hey": "hello"}, shared.Update{
		Id:      "123",
		Uid:     "a-uid",
		Applied: "2024-01-01Tsomething",
		Author:  "an-author",
		Type:    "a-type",
		Changes: []shared.Change{
			{Key: "a-key", Old: json.RawMessage("old"), New: json.RawMessage("new")},
		},
	})
	assert.Equal(t, errExpected, err)
}

func TestClientPut(t *testing.T) {
	dynamodbClient := newMockDynamodbClient(t)
	dynamodbClient.EXPECT().
		PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item: map[string]types.AttributeValue{
				"hey": &types.AttributeValueMemberS{Value: "hello"},
			},
		}).
		Return(nil, errExpected)

	client := &Client{
		svc:       dynamodbClient,
		tableName: tableName,
	}

	err := client.Put(ctx, map[string]string{"hey": "hello"})
	assert.Equal(t, errExpected, err)
}

func TestClientGet(t *testing.T) {
	dynamodbClient := newMockDynamodbClient(t)
	dynamodbClient.EXPECT().
		GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(tableName),
			Key: map[string]types.AttributeValue{
				"uid": &types.AttributeValueMemberS{Value: "my-uid"},
			},
		}).
		Return(&dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"uid":     &types.AttributeValueMemberS{Value: "my-uid"},
				"lpaType": &types.AttributeValueMemberS{Value: "property-and-affairs"},
			},
		}, nil)

	client := &Client{
		svc:       dynamodbClient,
		tableName: tableName,
	}

	lpa, err := client.Get(ctx, "my-uid")
	assert.Nil(t, err)
	assert.Equal(t, shared.Lpa{Uid: "my-uid", LpaInit: shared.LpaInit{LpaType: shared.LpaTypePropertyAndAffairs}}, lpa)
}

func TestClientGetWhenClientErrors(t *testing.T) {
	dynamodbClient := newMockDynamodbClient(t)
	dynamodbClient.EXPECT().
		GetItem(ctx, mock.Anything).
		Return(nil, errExpected)

	client := &Client{svc: dynamodbClient}

	_, err := client.Get(ctx, "my-uid")
	assert.Equal(t, errExpected, err)
}

func TestClientGetList(t *testing.T) {
	dynamodbClient := newMockDynamodbClient(t)
	dynamodbClient.EXPECT().
		BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
			RequestItems: map[string]types.KeysAndAttributes{
				tableName: {
					Keys: []map[string]types.AttributeValue{{
						"uid": &types.AttributeValueMemberS{Value: "my-uid"},
					}, {
						"uid": &types.AttributeValueMemberS{Value: "another-uid"},
					}},
				},
			},
		}).
		Return(&dynamodb.BatchGetItemOutput{
			Responses: map[string][]map[string]types.AttributeValue{
				tableName: {{
					"uid":     &types.AttributeValueMemberS{Value: "my-uid"},
					"lpaType": &types.AttributeValueMemberS{Value: "property-and-affairs"},
				}, {
					"uid":     &types.AttributeValueMemberS{Value: "another-uid"},
					"lpaType": &types.AttributeValueMemberS{Value: "personal-welfare"},
				}},
			},
		}, nil)

	client := &Client{
		svc:       dynamodbClient,
		tableName: tableName,
	}

	lpas, err := client.GetList(ctx, []string{"my-uid", "another-uid"})
	assert.Nil(t, err)
	assert.Equal(t, []shared.Lpa{
		{Uid: "my-uid", LpaInit: shared.LpaInit{LpaType: shared.LpaTypePropertyAndAffairs}},
		{Uid: "another-uid", LpaInit: shared.LpaInit{LpaType: shared.LpaTypePersonalWelfare}},
	}, lpas)
}

func TestClientGetListWhenClientErrors(t *testing.T) {
	dynamodbClient := newMockDynamodbClient(t)
	dynamodbClient.EXPECT().
		BatchGetItem(ctx, mock.Anything).
		Return(nil, errExpected)

	client := &Client{svc: dynamodbClient}

	_, err := client.GetList(ctx, []string{"my-uid", "another-uid"})
	assert.Equal(t, errExpected, err)
}

func TestClientGetChanges(t *testing.T) {
	dynamodbClient := newMockDynamodbClient(t)
	paginatorFactory := newMockPaginatorFactory(t)
	queryPaginator := newMockQueryPaginator(t)

	// First page
	page1 := &dynamodb.QueryOutput{
		Items: []map[string]types.AttributeValue{
			{
				"id":      &types.AttributeValueMemberS{Value: "1231"},
				"uid":     &types.AttributeValueMemberS{Value: "my-uid"},
				"applied": &types.AttributeValueMemberS{Value: "2024-01-01T00:00:00Z"},
				"author":  &types.AttributeValueMemberS{Value: "author-1"},
				"type":    &types.AttributeValueMemberS{Value: "type-1"},
				"changes": &types.AttributeValueMemberL{Value: []types.AttributeValue{
					&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
						"Key": &types.AttributeValueMemberS{Value: "key-1"},
						"Old": &types.AttributeValueMemberB{Value: []byte("old-1")},
						"New": &types.AttributeValueMemberB{Value: []byte("new-1")},
					}},
				}},
			},
		},
	}

	// Second page
	page2 := &dynamodb.QueryOutput{
		Items: []map[string]types.AttributeValue{
			{
				"id":      &types.AttributeValueMemberS{Value: "1232"},
				"uid":     &types.AttributeValueMemberS{Value: "my-uid"},
				"applied": &types.AttributeValueMemberS{Value: "2024-01-02T00:00:00Z"},
				"author":  &types.AttributeValueMemberS{Value: "author-2"},
				"type":    &types.AttributeValueMemberS{Value: "type-2"},
				"changes": &types.AttributeValueMemberL{Value: []types.AttributeValue{
					&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
						"Key": &types.AttributeValueMemberS{Value: "key-2"},
						"Old": &types.AttributeValueMemberB{Value: []byte("old-2")},
						"New": &types.AttributeValueMemberB{Value: []byte("new-2")},
					}},
				}},
			},
		},
	}

	s := "#0 = :0"
	scanIndexForward := false
	paginatorFactory.EXPECT().
		NewQueryPaginator(&dynamodb.QueryInput{
			TableName:                aws.String(changesTableName),
			ExpressionAttributeNames: map[string]string{"#0": "uid"},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":0": &types.AttributeValueMemberS{Value: "my-uid"},
			},
			KeyConditionExpression: &s,
			ScanIndexForward:       &scanIndexForward,
		}).
		Return(queryPaginator)

	queryPaginator.EXPECT().HasMorePages().Return(true).Once()
	queryPaginator.EXPECT().NextPage(ctx).Return(page1, nil).Once()
	queryPaginator.EXPECT().HasMorePages().Return(true).Once()
	queryPaginator.EXPECT().NextPage(ctx).Return(page2, nil).Once()
	queryPaginator.EXPECT().HasMorePages().Return(false).Once()

	client := &Client{
		svc:              dynamodbClient,
		changesTableName: changesTableName,
		paginatorFactory: paginatorFactory,
	}

	updates, err := client.GetChanges(ctx, "my-uid")
	assert.Nil(t, err)
	assert.Equal(t, []shared.Update{
		{
			Id:      "1231",
			Uid:     "my-uid",
			Applied: "2024-01-01T00:00:00Z",
			Author:  "author-1",
			Type:    "type-1",
			Changes: []shared.Change{
				{
					Key: "key-1",
					Old: json.RawMessage("old-1"),
					New: json.RawMessage("new-1"),
				},
			},
		},
		{
			Id:      "1232",
			Uid:     "my-uid",
			Applied: "2024-01-02T00:00:00Z",
			Author:  "author-2",
			Type:    "type-2",
			Changes: []shared.Change{
				{
					Key: "key-2",
					Old: json.RawMessage("old-2"),
					New: json.RawMessage("new-2"),
				},
			},
		},
	}, updates)
}

func TestClientGetChangesErrorOnQuery(t *testing.T) {
	dynamodbClient := newMockDynamodbClient(t)
	paginatorFactory := newMockPaginatorFactory(t)
	queryPaginator := newMockQueryPaginator(t)

	s := "#0 = :0"
	scanIndexForward := false
	paginatorFactory.EXPECT().
		NewQueryPaginator(&dynamodb.QueryInput{
			TableName:                aws.String(changesTableName),
			ExpressionAttributeNames: map[string]string{"#0": "uid"},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":0": &types.AttributeValueMemberS{Value: "my-uid"},
			},
			KeyConditionExpression: &s,
			ScanIndexForward:       &scanIndexForward,
		}).
		Return(queryPaginator)

	queryPaginator.EXPECT().HasMorePages().Return(true).Once()
	queryPaginator.EXPECT().NextPage(ctx).Return(nil, errExpected).Once()

	client := &Client{
		svc:              dynamodbClient,
		changesTableName: changesTableName,
		paginatorFactory: paginatorFactory,
	}

	updates, err := client.GetChanges(ctx, "my-uid")
	assert.Nil(t, updates)
	assert.Equal(t, errExpected, err)
}
