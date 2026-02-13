package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/google/uuid"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/ddb"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/event"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-go-common/telemetry"
)

type EventClient interface {
	SendLpaUpdated(ctx context.Context, event event.LpaUpdated, metric *event.Metric) error
}

type Logger interface {
	Error(string, ...any)
	Warn(string, ...any)
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
	environment string
	logger      Logger
	now         func() time.Time
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

	subject, _ := claims.GetSubject()
	update.Author = shared.URN(subject)

	redundantErrors, err := redundantChangeErrors(update.Changes)
	if err != nil {
		l.logger.Error("error evaluating redundant changes", slog.Any("err", err))
		return shared.ProblemInternalServerError.Respond()
	}

	if len(redundantErrors) > 0 {
		problem := shared.ProblemInvalidRequest
		problem.Errors = redundantErrors

		return problem.Respond()
	}

	applyable, errors := validateUpdate(update, &lpa)
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
	update.Applied = l.now().UTC().Format(time.RFC3339)

	if err := l.store.PutChanges(ctx, lpa, update); err != nil {
		l.logger.Error("error saving changes", slog.Any("err", err))
		return shared.ProblemInternalServerError.Respond()
	}

	var measureName string
	switch v := applyable.(type) {
	case AttorneySign:
		if v.Channel == shared.ChannelOnline {
			measureName = "ONLINEATTORNEY"
		} else {
			measureName = "PAPERATTORNEY"
		}
	case CertificateProviderSign:
		if v.Channel == shared.ChannelOnline {
			measureName = "ONLINECERTIFICATEPROVIDER"
		} else {
			measureName = "PAPERCERTIFICATEPROVIDER"
		}
	case TrustCorporationSign:
		if v.Channel == shared.ChannelOnline {
			measureName = "ONLINETRUSTCORPORATION"
		} else {
			measureName = "PAPERTRUSTCORPORATION"
		}
	}

	var metric *event.Metric
	if measureName != "" {
		metric = &event.Metric{
			Project:          "MRLPA",
			Category:         "metric",
			Subcategory:      "FunnelCompletionRate",
			Environment:      l.environment,
			MeasureName:      measureName,
			MeasureValue:     "1",
			MeasureValueType: "BIGINT",
			Time:             strconv.FormatInt(l.now().UnixMilli(), 10),
		}
	}

	body, err := json.Marshal(lpa)
	if err != nil {
		l.logger.Error("error marshalling LPA", slog.Any("err", err))
		return shared.ProblemInternalServerError.Respond()
	}

	if err := l.eventClient.SendLpaUpdated(ctx, event.LpaUpdated{
		Uid:        lpa.Uid,
		ChangeType: update.Type,
	}, metric); err != nil {
		l.logger.Error("unexpected error occurred", slog.Any("err", err))
	}

	response.StatusCode = 201
	response.Body = string(body)

	return response, nil
}

func main() {
	ctx := context.Background()
	logger := telemetry.NewLogger("opg-data-lpa-store/update")

	// set endpoint to "" outside dev to use default AWS resolver
	endpointURL := os.Getenv("AWS_BASE_URL")

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("failed to load aws config", slog.Any("err", err))
	}

	if endpointURL != "" {
		cfg.BaseEndpoint = aws.String(endpointURL)
	}

	l := &Lambda{
		eventClient: event.NewClient(cfg, os.Getenv("EVENT_BUS_NAME")),
		store: ddb.New(
			cfg,
			os.Getenv("DDB_TABLE_NAME_DEEDS"),
			os.Getenv("DDB_TABLE_NAME_CHANGES"),
		),
		verifier:    shared.NewJWTVerifier(cfg, logger),
		environment: os.Getenv("ENVIRONMENT"),
		logger:      logger,
		now:         time.Now,
	}

	lambda.Start(l.HandleEvent)
}
