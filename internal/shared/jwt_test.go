package shared

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

var secretKey = []byte("mysupersecrettestkeythatis128bits")

var verifier = JWTVerifier{
	secretKey: secretKey,
	Logger:    nil,
}

func createToken(claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, _ := token.SignedString(secretKey)

	return tokenString
}

func TestVerifyEmptyJwt(t *testing.T) {
	_, err := verifier.verifyToken("")
	assert.NotNil(t, err)
}

func TestVerifyExpInPast(t *testing.T) {
	token := createToken(jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * -24).Unix(),
		"iat": time.Now().Add(time.Hour * -24).Unix(),
		"iss": "opg.poas.makeregister",
		"sub": "urn:opg:poas:makeregister:users:e6707412-c9cd-4547-b428-7039a87e985e",
	})

	_, err := verifier.verifyToken(token)

	assert.NotNil(t, err)
	if err != nil {
		assert.Containsf(t, err.Error(), "token is expired", "")
	}
}

func TestVerifyIatInFuture(t *testing.T) {
	token := createToken(jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Add(time.Hour * 24).Unix(),
		"iss": "opg.poas.sirius",
		"sub": "urn:opg:sirius:users:34",
	})

	_, err := verifier.verifyToken(token)

	assert.NotNil(t, err)
	if err != nil {
		assert.Containsf(t, err.Error(), "IssuedAt must not be in the future", "")
	}
}

func TestVerifyIssuer(t *testing.T) {
	token := createToken(jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Add(time.Hour * -24).Unix(),
		"iss": "daadsdaadsadsads",
		"sub": "urn:opg:poas:makeregister:users:e6707412-c9cd-4547-b428-7039a87e985e",
	})

	_, err := verifier.verifyToken(token)

	assert.NotNil(t, err)
	if err != nil {
		assert.Containsf(t, err.Error(), "invalid issuer", "")
	}
}

func TestVerifySub(t *testing.T) {
	tests := map[string]struct {
		iss        string
		sub        string
		shouldFail bool
	}{
		"sirius empty":       {iss: "opg.poas.sirius", sub: "", shouldFail: true},
		"makeregister empty": {iss: "opg.poas.makeregister", sub: "", shouldFail: true},
		"use empty":          {iss: "opg.poas.use", sub: "", shouldFail: true},
		"valid sirius":       {iss: "opg.poas.sirius", sub: "urn:opg:sirius:users:34", shouldFail: false},
		"valid makeregister": {
			iss:        "opg.poas.makeregister",
			sub:        "urn:opg:poas:makeregister:users:e6707412-c9cd-4547-b428-7039a87e985e",
			shouldFail: false,
		},
		"valid use": {
			iss:        "opg.poas.use",
			sub:        "urn:opg:poas:use:users:ccba2c6c-33c6-497c-8248-25241ebf7edd",
			shouldFail: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			token := createToken(jwt.MapClaims{
				"exp": time.Now().Add(time.Hour * 24).Unix(),
				"iat": time.Now().Add(time.Hour * -24).Unix(),
				"iss": tc.iss,
				"sub": tc.sub,
			})

			_, err := verifier.verifyToken(token)

			if tc.shouldFail {
				assert.NotNil(t, err)
				if err != nil {
					assert.Containsf(t, err.Error(), "subject is not a valid URN", "")
				}
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestVerifyHeaderNoJWTHeader(t *testing.T) {
	event := events.APIGatewayProxyRequest{
		MultiValueHeaders: map[string][]string{},
	}

	_, err := verifier.VerifyHeader(event)
	assert.NotNil(t, err)
}

func TestVerifyHeader(t *testing.T) {
	token := createToken(jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Add(time.Hour * -24).Unix(),
		"iss": "opg.poas.sirius",
		"sub": "urn:opg:sirius:users:34",
	})

	event := events.APIGatewayProxyRequest{
		MultiValueHeaders: map[string][]string{
			"X-Jwt-Authorization": {
				fmt.Sprintf("Bearer %s", token),
			},
		},
	}

	logger := newMockLogger(t)
	logger.EXPECT().
		Info("JWT valid for urn:opg:sirius:users:34", slog.Any("subject", "urn:opg:sirius:users:34"))

	verifier.Logger = logger

	_, err := verifier.VerifyHeader(event)
	assert.Nil(t, err)
}
