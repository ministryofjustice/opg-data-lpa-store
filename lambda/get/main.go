package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ministryofjustice/opg-data-lpa-deed/lambda/shared"
	"github.com/ministryofjustice/opg-go-common/logging"
)

const test = "this should fail"

type Logger interface {
	Print(...interface{})
}

type Lambda struct {
	store  shared.Client
	logger Logger
}

func (l *Lambda) HandleEvent(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       "{\"code\":\"INTERNAL_SERVER_ERROR\",\"detail\":\"Internal server error\"}",
	}

	lpa, err := l.store.Get(ctx, event.PathParameters["uid"])

	if err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	body, err := json.Marshal(lpa)

	if err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	response.StatusCode = 200
	response.Body = string(body)

	return response, nil
}

func main() {
	l := &Lambda{
		store:  shared.NewDynamoDB(os.Getenv("DDB_TABLE_NAME_DEEDS")),
		logger: logging.New(os.Stdout, "opg-data-lpa-deed"),
	}

	lambda.Start(l.HandleEvent)
}
