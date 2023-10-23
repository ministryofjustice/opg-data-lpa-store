package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-openapi/jsonpointer"
	"github.com/ministryofjustice/opg-data-lpa-deed/lambda/shared"
	"github.com/ministryofjustice/opg-go-common/logging"
)

type Logger interface {
	Print(...interface{})
}

type Lambda struct {
	store  shared.Client
	logger Logger
}

func (l *Lambda) HandleEvent(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var update shared.Update
	response := events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       "{\"code\":\"INTERNAL_SERVER_ERROR\",\"detail\":\"Internal server error\"}",
	}

	err := json.Unmarshal([]byte(event.Body), &update)
	if err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	lpa, err := l.store.Get(ctx, event.PathParameters["uid"])
	if err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	err = applyUpdate(&lpa, update)
	if err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	err = l.store.Put(ctx, lpa)
	if err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	body, err := json.Marshal(lpa)

	if err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	response.StatusCode = 201
	response.Body = string(body)

	return response, nil
}

func applyUpdate(lpa *shared.Lpa, update shared.Update) error {
	for _, change := range update.Changes {
		pointer, err := jsonpointer.New(change.Key)
		if err != nil {
			return err
		}

		current, _, err := pointer.Get(*lpa)
		if err != nil {
			return err
		}

		if current != change.Old {
			err = fmt.Errorf("existing value for %s does not match request", change.Key)
			return err
		}

		_, err = pointer.Set(lpa, change.New)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	l := &Lambda{
		store:  shared.NewDynamoDB(os.Getenv("DDB_TABLE_NAME_DEEDS")),
		logger: logging.New(os.Stdout, "opg-data-lpa-deed"),
	}

	lambda.Start(l.HandleEvent)
}
