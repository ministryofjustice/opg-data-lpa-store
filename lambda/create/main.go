package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ministryofjustice/opg-data-lpa-deed/lambda/shared"
	"github.com/ministryofjustice/opg-go-common/logging"
)

type Response struct {
}

type Logger interface {
	Print(...interface{})
}

type Lambda struct {
	store  shared.Client
	logger Logger
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

	// save
	err = l.store.Put(ctx, data)

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
	l := &Lambda{
		store:  shared.NewDynamoDB(os.Getenv("DDB_TABLE_NAME_DEEDS")),
		logger: logging.New(os.Stdout, "opg-data-lpa-deed"),
	}

	lambda.Start(l.HandleEvent)
}
