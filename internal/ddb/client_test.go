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

var (
	ctx              = context.WithValue(context.Background(), "for", "testing")
	tableName        = "a-table"
	changesTableName = "a-change-table"
	expectedError    = errors.New("hey")
)

func TestNew(t *testing.T) {
	cfg := aws.Config{Region: "somewhere"}

	client := New(cfg, tableName, changesTableName)
	assert.IsType(t, (*dynamodb.Client)(nil), client.ddb)
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
						"uid":     &types.AttributeValueMemberS{Value: "a-uid"},
						"applied": &types.AttributeValueMemberS{Value: "2024-01-01Tsomething"},
						"author":  &types.AttributeValueMemberS{Value: "an-author"},
						"type":    &types.AttributeValueMemberS{Value: "a-type"},
						"change": &types.AttributeValueMemberL{Value: []types.AttributeValue{
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
		Return(nil, expectedError)

	client := &Client{
		ddb:              dynamodbClient,
		tableName:        tableName,
		changesTableName: changesTableName,
	}

	err := client.PutChanges(ctx, map[string]string{"hey": "hello"}, shared.Update{
		Uid:     "a-uid",
		Applied: "2024-01-01Tsomething",
		Author:  "an-author",
		Type:    "a-type",
		Changes: []shared.Change{
			{Key: "a-key", Old: json.RawMessage("old"), New: json.RawMessage("new")},
		},
	})
	assert.Equal(t, expectedError, err)
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
		Return(nil, expectedError)

	client := &Client{
		ddb:       dynamodbClient,
		tableName: tableName,
	}

	err := client.Put(ctx, map[string]string{"hey": "hello"})
	assert.Equal(t, expectedError, err)
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
		ddb:       dynamodbClient,
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
		Return(nil, expectedError)

	client := &Client{ddb: dynamodbClient}

	_, err := client.Get(ctx, "my-uid")
	assert.Equal(t, expectedError, err)
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
				tableName: []map[string]types.AttributeValue{{
					"uid":     &types.AttributeValueMemberS{Value: "my-uid"},
					"lpaType": &types.AttributeValueMemberS{Value: "property-and-affairs"},
				}, {
					"uid":     &types.AttributeValueMemberS{Value: "another-uid"},
					"lpaType": &types.AttributeValueMemberS{Value: "personal-welfare"},
				}},
			},
		}, nil)

	client := &Client{
		ddb:       dynamodbClient,
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
		Return(nil, expectedError)

	client := &Client{ddb: dynamodbClient}

	_, err := client.GetList(ctx, []string{"my-uid", "another-uid"})
	assert.Equal(t, expectedError, err)
}
