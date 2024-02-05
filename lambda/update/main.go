package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/ddb"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-go-common/logging"
)

type Logger interface {
	Print(...interface{})
}

type Store interface {
	PutChanges(ctx context.Context, data any, update shared.Update) error
	Get(ctx context.Context, uid string) (shared.Lpa, error)
}

type Verifier interface {
	VerifyHeader(events.APIGatewayProxyRequest) (*shared.LpaStoreClaims, error)
}

type Lambda struct {
	store    Store
	verifier Verifier
	logger   Logger
}

func (l *Lambda) HandleEvent(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	claims, err := l.verifier.VerifyHeader(event)
	if err != nil {
		l.logger.Print("Unable to verify JWT from header")
		return shared.ProblemUnauthorisedRequest.Respond()
	}

	l.logger.Print("Successfully parsed JWT from event header")

	response := events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       "{\"code\":\"INTERNAL_SERVER_ERROR\",\"detail\":\"Internal server error\"}",
	}

	var update shared.Update
	if err = json.Unmarshal([]byte(event.Body), &update); err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}

	lpa, err := l.store.Get(ctx, event.PathParameters["uid"])
	if err != nil {
		l.logger.Print(err)
		return shared.ProblemInternalServerError.Respond()
	}
	if lpa.Uid == "" {
		l.logger.Print("Uid not found")
		return shared.ProblemNotFoundRequest.Respond()
	}

	applyable, errors := validateUpdate(update)
	if len(errors) > 0 {
		problem := shared.ProblemInvalidRequest
		problem.Errors = errors

		return problem.Respond()
	}

	if errors := applyable.Apply(&lpa); len(errors) > 0 {
		problem := shared.ProblemInvalidRequest
		problem.Errors = errors

		return problem.Respond()
	}

	update.Id = uuid.NewString()
	update.Uid = lpa.Uid
	update.Applied = time.Now().Format(time.RFC3339)
	update.Author, _ = claims.GetSubject()

	if err := l.store.PutChanges(ctx, lpa, update); err != nil {
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

func main() {
	l := &Lambda{
		store:    ddb.New(
			os.Getenv("AWS_DYNAMODB_ENDPOINT"),
			os.Getenv("DDB_TABLE_NAME_DEEDS"),
			os.Getenv("DDB_TABLE_NAME_CHANGES"),
		),
		verifier: shared.NewJWTVerifier(),
		logger:   logging.New(os.Stdout, "opg-data-lpa-store"),
	}

	lambda.Start(l.HandleEvent)
}
