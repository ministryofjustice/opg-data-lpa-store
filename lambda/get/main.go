package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/ddb"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-go-common/telemetry"
)

type Logger interface {
	Error(string, ...any)
	Info(string, ...any)
	Debug(string, ...any)
}

type Store interface {
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
	_, err := l.verifier.VerifyHeader(event)
	if err != nil {
		l.logger.Info("Unable to verify JWT from header")
		return shared.ProblemUnauthorisedRequest.Respond()
	}

	l.logger.Debug("Successfully parsed JWT from event header")

	response := events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       "{\"code\":\"INTERNAL_SERVER_ERROR\",\"detail\":\"Internal server error\"}",
	}

	lpa, err := l.store.Get(ctx, event.PathParameters["uid"])

	// If item can't be found in DynamoDB then it returns empty object hence 404 error returned if
	// empty object returned
	if lpa.Uid == "" {
		l.logger.Debug("Uid not found")
		return shared.ProblemNotFoundRequest.Respond()
	}

	if err != nil {
		l.logger.Error("error fetching LPA", slog.Any("err", err))
		return shared.ProblemInternalServerError.Respond()
	}

	body, err := json.Marshal(lpa)

	if err != nil {
		l.logger.Error("error marshalling LPA", slog.Any("err", err))
		return shared.ProblemInternalServerError.Respond()
	}

	log.Println(string(body))

	response.StatusCode = 200
	response.Body = string(body)

	return response, nil
}

func main() {
	ctx := context.Background()
	logger := telemetry.NewLogger("opg-data-lpa-store/get")

	// set endpoint to "" outside dev to use default AWS resolver
	endpointURL := os.Getenv("AWS_BASE_URL")

	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
		if endpointURL != "" {
			o.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{URL: endpointURL, HostnameImmutable: true}, nil
				},
			)
		}

		return nil
	})
	if err != nil {
		logger.Error("failed to load aws config", slog.Any("err", err))
	}

	l := &Lambda{
		store: ddb.New(
			cfg,
			os.Getenv("DDB_TABLE_NAME_DEEDS"),
			os.Getenv("DDB_TABLE_NAME_CHANGES"),
		),
		verifier: shared.NewJWTVerifier(cfg, logger),
		logger:   logger,
	}

	lambda.Start(l.HandleEvent)
}
