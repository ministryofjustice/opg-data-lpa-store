package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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

	lpa, err := l.store.Get(ctx, event.PathParameters["uid"])

    // If item can't be found in DynamoDB then it returns empty object hence 404 error returned if
    // empty object returned
	if lpa.Uid == "" {
		l.logger.Print("Uid not found")
		return shared.ProblemNotFoundRequest.Respond()
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
		store:    shared.NewDynamoDB(os.Getenv("DDB_TABLE_NAME_DEEDS")),
		verifier: shared.NewJWTVerifier(),
		logger:   logging.New(os.Stdout, "opg-data-lpa-store"),
	}

	lambda.Start(l.HandleEvent)
}
