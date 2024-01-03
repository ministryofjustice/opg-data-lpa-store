package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-openapi/jsonpointer"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/ddb"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-go-common/logging"
)

type Logger interface {
	Print(...interface{})
}

type Lambda struct {
	store    shared.Client
	verifier shared.JWTVerifier
	logger   Logger
}

func (l *Lambda) HandleEvent(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !l.verifier.VerifyHeader(event) {
		l.logger.Print("Unable to verify JWT from header")
		return shared.ProblemUnauthorisedRequest.Respond()
	}

	l.logger.Print("Successfully parsed JWT from event header")

	response := events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       "{\"code\":\"INTERNAL_SERVER_ERROR\",\"detail\":\"Internal server error\"}",
	}

	var update shared.Update
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

	validationErrs, err := applyUpdate(&lpa, update)

	if err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	if len(validationErrs) > 0 {
		problem := shared.ProblemInvalidRequest
		problem.Errors = validationErrs

		return problem.Respond()
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

func applyUpdate(lpa *shared.Lpa, update shared.Update) ([]shared.FieldError, error) {
	validationErrs := []shared.FieldError{}

	for index, change := range update.Changes {
		pointer, err := jsonpointer.New(change.Key)
		if err != nil {
			return validationErrs, err
		}

		current, _, err := pointer.Get(*lpa)
		if err != nil {
			return validationErrs, err
		}

		if current != change.Old {
			validationErrs = append(validationErrs, shared.FieldError{
				Source: fmt.Sprintf("/changes/%d/old", index),
				Detail: "does not match existing value",
			})
		}

		_, err = pointer.Set(lpa, change.New)
		if err != nil {
			return validationErrs, err
		}
	}

	return validationErrs, nil
}

func main() {
	l := &Lambda{
		store:    ddb.New(os.Getenv("AWS_DYNAMODB_ENDPOINT"), os.Getenv("DDB_TABLE_NAME_DEEDS")),
		verifier: shared.NewJWTVerifier(),
		logger:   logging.New(os.Stdout, "opg-data-lpa-store"),
	}

	lambda.Start(l.HandleEvent)
}
