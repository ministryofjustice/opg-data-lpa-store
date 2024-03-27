package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/ddb"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/event"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/objectstore"
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
	Put(ctx context.Context, data any) error
	Get(ctx context.Context, uid string) (shared.Lpa, error)
}

type S3Client interface {
	Put(ctx context.Context, objectKey string, obj any) error
}

type Verifier interface {
	VerifyHeader(events.APIGatewayProxyRequest) (*shared.LpaStoreClaims, error)
}

type Lambda struct {
	eventClient      EventClient
	staticLpaStorage S3Client
	store            Store
	verifier         Verifier
	logger           Logger
}

func (l *Lambda) HandleEvent(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, err := l.verifier.VerifyHeader(req)
	if err != nil {
		l.logger.Info("Unable to verify JWT from header")
		return shared.ProblemUnauthorisedRequest.Respond()
	}

	l.logger.Debug("Successfully parsed JWT from event header")

	var input shared.LpaInit
	uid := req.PathParameters["uid"]

	response := events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       "{\"code\":\"INTERNAL_SERVER_ERROR\",\"detail\":\"Internal server error\"}",
	}

	// check for existing Lpa
	var existingLpa shared.Lpa
	existingLpa, err = l.store.Get(ctx, uid)
	if err != nil {
		l.logger.Error("error fetching LPA", slog.Any("err", err))
		return shared.ProblemInternalServerError.Respond()
	}

	if existingLpa.Uid == uid {
		problem := shared.ProblemInvalidRequest
		problem.Detail = "LPA with UID already exists"
		return problem.Respond()
	}

	err = json.Unmarshal([]byte(req.Body), &input)
	if err != nil {
		l.logger.Error("error unmarshalling request", slog.Any("err", err))
		return shared.ProblemInternalServerError.Respond()
	}

	// validation
	if errs := Validate(input); len(errs) > 0 {
		problem := shared.ProblemInvalidRequest
		problem.Errors = errs

		return problem.Respond()
	}

	data := shared.Lpa{LpaInit: input}
	data.Uid = uid
	data.Status = shared.LpaStatusProcessing
	data.UpdatedAt = time.Now()

	// save
	if err = l.store.Put(ctx, data); err != nil {
		l.logger.Error("error saving LPA", slog.Any("err", err))
		return shared.ProblemInternalServerError.Respond()
	}

	// save to static storage as JSON
	objectKey := fmt.Sprintf("%s/donor-executed-lpa.json", data.Uid)

	if err = l.staticLpaStorage.Put(ctx, objectKey, data); err != nil {
		l.logger.Error("error saving static record", slog.Any("err", err))
		return shared.ProblemInternalServerError.Respond()
	}

	// send lpa-updated event
	err = l.eventClient.SendLpaUpdated(ctx, event.LpaUpdated{
		Uid:        uid,
		ChangeType: "CREATE",
	})

	if err != nil {
		l.logger.Error("unexpected error occurred", slog.Any("err", err))
	}

	// respond
	response.StatusCode = 201
	response.Body = `{}`

	return response, nil
}

func main() {
	logger := telemetry.NewLogger("opg-data-lpa-store/create")

	// set endpoint to "" outside dev to use default AWS resolver
	endpointURL := os.Getenv("AWS_S3_ENDPOINT")

	var endpointResolverWithOptions aws.EndpointResolverWithOptions
	if endpointURL != "" {
		endpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: endpointURL, HostnameImmutable: true}, nil
			},
		)
	}

	ctx := context.Background()

	eventClientConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("Failed to load event client configuration", slog.Any("err", err))
	}

	s3Config, err := config.LoadDefaultConfig(
		ctx,
		func(o *config.LoadOptions) error {
			o.EndpointResolverWithOptions = endpointResolverWithOptions
			return nil
		},
	)
	if err != nil {
		logger.Error("Failed to load S3 configuration", slog.Any("err", err))
	}

	l := &Lambda{
		eventClient: event.NewClient(eventClientConfig, os.Getenv("EVENT_BUS_NAME")),
		store: ddb.New(
			os.Getenv("AWS_DYNAMODB_ENDPOINT"),
			os.Getenv("DDB_TABLE_NAME_DEEDS"),
			os.Getenv("DDB_TABLE_NAME_CHANGES"),
		),
		staticLpaStorage: objectstore.NewS3Client(
			s3Config,
			os.Getenv("S3_BUCKET_NAME_ORIGINAL"),
		),
		verifier: shared.NewJWTVerifier(logger),
		logger:   logger,
	}

	lambda.Start(l.HandleEvent)
}
