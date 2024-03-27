package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/google/uuid"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/ddb"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/event"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-go-common/telemetry"
)

type EventClient interface {
	SendLpaUpdated(ctx context.Context, event event.LpaUpdated) error
}

type Logger interface {
	Error(string, ...any)
	Info(string, ...any)
	Debug(string, ...any)
}

type Store interface {
	PutChanges(ctx context.Context, data any, update shared.Update) error
	Get(ctx context.Context, uid string) (shared.Lpa, error)
}

type Verifier interface {
	VerifyHeader(events.APIGatewayProxyRequest) (*shared.LpaStoreClaims, error)
}

type Lambda struct {
	eventClient EventClient
	store       Store
	verifier    Verifier
	logger      Logger
}

func (l *Lambda) HandleEvent(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	claims, err := l.verifier.VerifyHeader(req)
	if err != nil {
		l.logger.Info("Unable to verify JWT from header")
		return shared.ProblemUnauthorisedRequest.Respond()
	}

	l.logger.Debug("Successfully parsed JWT from event header")

	response := events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       "{\"code\":\"INTERNAL_SERVER_ERROR\",\"detail\":\"Internal server error\"}",
	}

	var update shared.Update
	if err = json.Unmarshal([]byte(req.Body), &update); err != nil {
		l.logger.Error("error unmarshalling request", slog.Any("err", err))
		return shared.ProblemInternalServerError.Respond()
	}

	lpa, err := l.store.Get(ctx, req.PathParameters["uid"])
	if err != nil {
		l.logger.Error("error fetching LPA", slog.Any("err", err))
		return shared.ProblemInternalServerError.Respond()
	}
	if lpa.Uid == "" {
		l.logger.Debug("Uid not found")
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
		l.logger.Error("error saving changes", slog.Any("err", err))
		return shared.ProblemInternalServerError.Respond()
	}

	body, err := json.Marshal(lpa)
	if err != nil {
		l.logger.Error("error marshalling LPA", slog.Any("err", err))
		return shared.ProblemInternalServerError.Respond()
	}

	// send lpa-updated event
	err = l.eventClient.SendLpaUpdated(ctx, event.LpaUpdated{
		Uid:        lpa.Uid,
		ChangeType: update.Type,
	})

	if err != nil {
		l.logger.Error("unexpected error occurred", slog.Any("err", err))
	}

	response.StatusCode = 201
	response.Body = string(body)

	return response, nil
}

func main() {
	logger := telemetry.NewLogger("opg-data-lpa-store/update")
	ctx := context.Background()
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("Failed to load configuration", slog.Any("err", err))
	}

	l := &Lambda{
		eventClient: event.NewClient(awsConfig, os.Getenv("EVENT_BUS_NAME")),
		store: ddb.New(
			os.Getenv("AWS_DYNAMODB_ENDPOINT"),
			os.Getenv("DDB_TABLE_NAME_DEEDS"),
			os.Getenv("DDB_TABLE_NAME_CHANGES"),
		),
		verifier: shared.NewJWTVerifier(logger),
		logger:   logger,
	}

	lambda.Start(l.HandleEvent)
}
