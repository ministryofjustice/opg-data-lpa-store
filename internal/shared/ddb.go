package shared

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-xray-sdk-go/xray"
)

type DynamoDBClient struct {
	ddb       *dynamodb.DynamoDB
	tableName string
}

func (c DynamoDBClient) Put(ctx context.Context, data Lpa) error {
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

func (c DynamoDBClient) Get(ctx context.Context, uid string) (Lpa, error) {
	lpa := Lpa{}

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

func NewDynamoDB(tableName string) DynamoDBClient {
	sess := session.Must(session.NewSession())

	endpoint := os.Getenv("AWS_DYNAMODB_ENDPOINT")
	sess.Config.Endpoint = &endpoint

	c := DynamoDBClient{
		ddb:       dynamodb.New(sess),
		tableName: tableName,
	}

	xray.AWS(c.ddb.Client)

	return c
}
