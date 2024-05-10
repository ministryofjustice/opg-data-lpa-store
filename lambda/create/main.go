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
	UploadFile(ctx context.Context, image shared.FileUpload, path string) (shared.File, error)
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
		if input.Channel == shared.ChannelPaper {
			l.logger.Info("encountered validation errors in lpa", slog.Any("uid", uid))
		} else {
			problem := shared.ProblemInvalidRequest
			problem.Errors = errs

			return problem.Respond()
		}
	}

	data := shared.Lpa{LpaInit: input}
	data.Uid = uid
	data.Status = shared.LpaStatusProcessing
	data.UpdatedAt = time.Now()

	if data.Channel == shared.ChannelPaper && len(input.RestrictionsAndConditionsImages) > 0 {
		data.RestrictionsAndConditionsImages = make([]shared.File, len(input.RestrictionsAndConditionsImages))
		for i, image := range input.RestrictionsAndConditionsImages {
			path := fmt.Sprintf("%s/scans/rc_%d_%s", data.Uid, i, image.Filename)

			data.RestrictionsAndConditionsImages[i], err = l.staticLpaStorage.UploadFile(ctx, image, path)

			if err != nil {
				l.logger.Error("error saving restrictions and conditions image", slog.Any("err", err))
				return shared.ProblemInternalServerError.Respond()
			}
		}
	}

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
	ctx := context.Background()
	logger := telemetry.NewLogger("opg-data-lpa-store/create")

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
		eventClient: event.NewClient(cfg, os.Getenv("EVENT_BUS_NAME")),
		store: ddb.New(
			cfg,
			os.Getenv("DDB_TABLE_NAME_DEEDS"),
			os.Getenv("DDB_TABLE_NAME_CHANGES"),
		),
		staticLpaStorage: objectstore.NewS3Client(
			cfg,
			os.Getenv("S3_BUCKET_NAME_ORIGINAL"),
		),
		verifier: shared.NewJWTVerifier(cfg, logger),
		logger:   logger,
	}

	lambda.Start(l.HandleEvent)
}
