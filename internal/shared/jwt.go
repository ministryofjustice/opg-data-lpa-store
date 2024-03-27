package shared

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	jwt "github.com/golang-jwt/jwt/v5"
	urn "github.com/leodido/go-urn"
)

const (
	sirius string = "opg.poas.sirius"
	mrlpa         = "opg.poas.makeregister"
)

var validIssuers []string = []string{
	sirius,
	mrlpa,
}

type LpaStoreClaims struct {
	jwt.RegisteredClaims
}

// note that default validation for RegisteredClaims checks exp is in the future
func (l LpaStoreClaims) Validate() error {
	// validate issued at (iat)
	iat, err := l.GetIssuedAt()
	if err != nil {
		return err
	}

	if iat.Time.After(time.Now()) {
		return errors.New("IssuedAt must not be in the future")
	}

	// validate issuer (iss)
	iss, err := l.GetIssuer()
	if err != nil {
		return err
	}

	isValid := false
	for _, validIssuer := range validIssuers {
		if validIssuer == iss {
			isValid = true
			break
		}
	}

	if !isValid {
		return errors.New("Invalid Issuer")
	}

	// validate subject (sub) depending on the issuer value
	sub, err := l.GetSubject()
	if err != nil {
		return err
	}

	_, isUrn := urn.Parse([]byte(sub))

	if !isUrn {
		switch iss {
		case mrlpa:
			return errors.New("Subject is not a valid URN")
		case sirius:
			emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
			if !emailRegex.MatchString(sub) {
				return errors.New("Subject is not a valid email or URN")
			}
		}
	}

	return nil
}

type JWTVerifier struct {
	secretKey []byte
}

type logger interface {
	Error(string, ...any)
}

func NewJWTVerifier(logger logger) JWTVerifier {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		logger.Error("Failed to load secretsmanager configuration", slog.Any("err", err))
	}

	client := secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
		o.BaseEndpoint = aws.String(os.Getenv("AWS_DYNAMODB_ENDPOINT"))
	})

	secretKey, err := client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(os.Getenv("JWT_SECRET_KEY_ID")),
	})

	if err != nil {
		logger.Error("Failed to fetch JWT signing secret", slog.Any("err", err))
	}

	return JWTVerifier{
		secretKey: []byte(*secretKey.SecretString),
	}
}

var bearerRegexp = regexp.MustCompile("^Bearer[ ]+")

// verify JWT from event header
// returns true if verified, false otherwise
func (v JWTVerifier) VerifyHeader(event events.APIGatewayProxyRequest) (*LpaStoreClaims, error) {
	jwtHeaders := GetEventHeader("X-Jwt-Authorization", event)

	if len(jwtHeaders) < 1 {
		return nil, fmt.Errorf("Invalid X-Jwt-Authorization header")
	}

	tokenStr := bearerRegexp.ReplaceAllString(jwtHeaders[0], "")
	claims, err := v.verifyToken(tokenStr)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

// tokenStr is the JWT token, minus any "Bearer: " prefix
func (v JWTVerifier) verifyToken(tokenStr string) (*LpaStoreClaims, error) {
	lsc := LpaStoreClaims{}

	parsedToken, err := jwt.ParseWithClaims(tokenStr, &lsc, func(token *jwt.Token) (interface{}, error) {
		return v.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("Invalid JWT")
	}

	return &lsc, nil
}
