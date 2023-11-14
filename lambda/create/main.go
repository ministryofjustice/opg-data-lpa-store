package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/shared"
	"github.com/ministryofjustice/opg-go-common/logging"
)

type Response struct {
}

type Logger interface {
	Print(...interface{})
}

type Lambda struct {
	store    shared.Client
	verifier shared.JWTVerifier
	logger   Logger
}

func (l *Lambda) HandleEvent(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var data shared.Lpa
	var err error

	response := events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       "{\"code\":\"INTERNAL_SERVER_ERROR\",\"detail\":\"Internal server error\"}",
	}

	err = l.verifier.VerifyHeader(event)
	if err == nil {
		l.logger.Print("Successfully parsed JWT from event header")
	} else {
		l.logger.Print(err)
	}

	err = json.Unmarshal([]byte(event.Body), &data)
	if err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	data.Uid = event.PathParameters["uid"]

	if data.Version == "" {
		problem := shared.ProblemInvalidRequest
		problem.Errors = []shared.FieldError{
			{Source: "/version", Detail: "must supply a valid version"},
		}

		return problem.Respond()
	}

	// check for existing Lpa
	var existingLpa shared.Lpa
	existingLpa, err = l.store.Get(ctx, data.Uid)
	if err != nil {
		return shared.ProblemInternalServerError.Respond()
	}
	if existingLpa.Uid == data.Uid {
		problem := shared.ProblemInvalidRequest
		problem.Detail = "LPA with UID already exists"
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
		store:    shared.NewDynamoDB(os.Getenv("DDB_TABLE_NAME_DEEDS")),
		verifier: shared.NewJWTVerifier(),
		logger:   logging.New(os.Stdout, "opg-data-lpa-store"),
	}

	lambda.Start(l.HandleEvent)
}
