package getstatic

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/objectstore"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-go-common/telemetry"
)

type Logger interface {
	Error(string, ...any)
	Info(string, ...any)
	Debug(string, ...any)
}

type Verifier interface {
	VerifyHeader(events.APIGatewayProxyRequest) (*shared.LpaStoreClaims, error)
}

type S3Client interface {
	Get(ctx context.Context, objectKey string) (string, error)
}

type Lambda struct {
	logger           Logger
	staticLpaStorage S3Client
	verifier         Verifier
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

	objectKey := fmt.Sprintf("%s/donor-executed-lpa.json", event.PathParameters["uid"])
	data, err := l.staticLpaStorage.Get(ctx, objectKey)
	if err != nil {
		l.logger.Error("error fetching static LPA", slog.Any("err", err))
		return shared.ProblemInternalServerError.Respond()
	}

	response.StatusCode = 200
	response.Body = data

	return response, nil
}

func main() {
	ctx := context.Background()
	logger := telemetry.NewLogger("opg-data-lpa-store/getstatic")

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
		logger: logger,
		staticLpaStorage: objectstore.NewS3Client(
			cfg,
			os.Getenv("S3_BUCKET_NAME_ORIGINAL"),
		),
		verifier: shared.NewJWTVerifier(cfg, logger),
	}

	lambda.Start(l.HandleEvent)
}
