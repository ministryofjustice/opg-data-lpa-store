package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/ministryofjustice/opg-data-lpa-deed/lambda/shared"
	"github.com/ministryofjustice/opg-go-common/logging"
)

type Response struct {
}

type Logger interface {
	Print(...interface{})
}

type Lambda struct {
	ddb       *dynamodb.DynamoDB
	tableName string
	logger    Logger
}

func (l *Lambda) HandleEvent(ctx context.Context, event events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	var data shared.Case
	response := events.LambdaFunctionURLResponse{
		StatusCode: 500,
		Body:       "{\"code\":\"INTERNAL_SERVER_ERROR\",\"detail\":\"Internal server error\"}",
	}

	err := json.Unmarshal([]byte(event.Body), &data)
	if err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	if data.Version == "" {
		problem := shared.ProblemInvalidRequest
		problem.Errors = []shared.FieldError{
			{Source: "/version", Detail: "must supply a valid version"},
		}

		return problem.Respond()
	}

	data.UpdatedAt = time.Now()

	// save to dynamodb
	item, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	_, err = l.ddb.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(l.tableName),
		Item:      item,
	})

	if err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	// respond
	body, err := json.Marshal(Response{})

	if err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	response.StatusCode = 201
	response.Body = string(body)

	return response, nil
}

func main() {
	sess := session.Must(session.NewSession())

	endpoint := os.Getenv("AWS_DYNAMODB_ENDPOINT")
	sess.Config.Endpoint = &endpoint

	l := &Lambda{
		ddb:       dynamodb.New(sess),
		tableName: os.Getenv("DDB_TABLE_NAME_DEEDS"),
		logger:    logging.New(os.Stdout, "opg-data-lpa-deed"),
	}

	xray.AWS(l.ddb.Client)

	lambda.Start(l.HandleEvent)
}
