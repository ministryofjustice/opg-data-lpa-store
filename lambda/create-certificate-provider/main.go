package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/ddb"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
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
	if !l.verifier.VerifyHeader(event) {
		l.logger.Print("Unable to verify JWT from header")
		return shared.ProblemUnauthorisedRequest.Respond()
	}

	l.logger.Print("Successfully parsed JWT from event header")

	uid := event.PathParameters["uid"]

	response := events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       "{\"code\":\"INTERNAL_SERVER_ERROR\",\"detail\":\"Internal server error\"}",
	}

	// check for existing Lpa
	existingLpa, err := l.store.Get(ctx, uid)
	if err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}
	if existingLpa.Uid == "" {
		return shared.ProblemNotFoundRequest.Respond()
	}

	var input CertificateProvider
	if err := json.Unmarshal([]byte(event.Body), &input); err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	// validation
	errors := Validate(input)
	if len(errors) > 0 {
		problem := shared.ProblemInvalidRequest
		problem.Errors = errors

		return problem.Respond()
	}

	input.UpdatedAt = time.Now()

	// save
	if err := l.store.Put(ctx, input); err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	// respond
	response.StatusCode = 201
	response.Body = `{}`

	return response, nil
}

func main() {
	l := &Lambda{
		store:    ddb.New(os.Getenv("AWS_DYNAMODB_ENDPOINT"), os.Getenv("DDB_TABLE_NAME_DEEDS")),
		verifier: shared.NewJWTVerifier(),
		logger:   logging.New(os.Stdout, "opg-data-lpa-store"),
	}

	lambda.Start(l.HandleEvent)
}
