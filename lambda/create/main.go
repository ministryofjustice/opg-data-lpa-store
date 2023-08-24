package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/ministryofjustice/opg-data-lpa-deed/lambda/shared"
	"github.com/ministryofjustice/opg-go-common/logging"
)

type Response struct {
}

type Logger interface {
	Print(...interface{})
}

type Lambda struct {
	ddb       dynamodbiface.DynamoDBAPI
	tableName string
	logger    Logger
}

func (l *Lambda) HandleEvent(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var data shared.Case
	log.Print(event)
	response := events.APIGatewayProxyResponse{
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

	_, err = l.ddb.PutItem(&dynamodb.PutItemInput{
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
		tableName: "deeds",
		logger:    logging.New(os.Stdout, "opg-data-lpa-deed"),
	}

	lambda.Start(l.HandleEvent)
}
