package shared

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

var secretKey = []byte("secret")

var verifier = JWTVerifier{
	secretKey: secretKey,
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
		"sub": "someone@someplace.somewhere.com",
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
		"sub": "someone@someplace.somewhere.com",
	})

	_, err := verifier.verifyToken(token)

	assert.NotNil(t, err)
	if err != nil {
		assert.Containsf(t, err.Error(), "Invalid Issuer", "")
	}
}

func TestVerifyBadSubForSiriusIssuer(t *testing.T) {
	token := createToken(jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Add(time.Hour * -24).Unix(),
		"iss": "opg.poas.sirius",
		"sub": "",
	})

	_, err := verifier.verifyToken(token)

	assert.NotNil(t, err)
	if err != nil {
		assert.Containsf(t, err.Error(), "Subject is not a valid email or URN", "")
	}
}

func TestVerifyBadSubForMRLPAIssuer(t *testing.T) {
	token := createToken(jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Add(time.Hour * -24).Unix(),
		"iss": "opg.poas.makeregister",
		"sub": "",
	})

	_, err := verifier.verifyToken(token)

	assert.NotNil(t, err)
	if err != nil {
		assert.Containsf(t, err.Error(), "Subject is not a valid URN", "")
	}
}

func TestVerifyGoodJwtSiriusSubs(t *testing.T) {
	token := createToken(jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Add(time.Hour * -24).Unix(),
		"iss": "opg.poas.sirius",
		"sub": "someone@someplace.somewhere.com",
	})

	_, err := verifier.verifyToken(token)
	assert.Nil(t, err)

	token = createToken(jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Add(time.Hour * -24).Unix(),
		"iss": "opg.poas.sirius",
		"sub": "urn:opg:sirius:users:34",
	})

	_, err = verifier.verifyToken(token)
	assert.Nil(t, err)
}

func TestVerifyGoodJwtMRLPASubs(t *testing.T) {
	token := createToken(jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Add(time.Hour * -24).Unix(),
		"iss": "opg.poas.makeregister",
		"sub": "urn:opg:poas:makeregister:users:e6707412-c9cd-4547-b428-7039a87e985e",
	})

	_, err := verifier.verifyToken(token)
	assert.Nil(t, err)
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
		"sub": "someone@someplace.somewhere.com",
	})

	event := events.APIGatewayProxyRequest{
		MultiValueHeaders: map[string][]string{
			"X-Jwt-Authorization": []string{
				fmt.Sprintf("Bearer %s", token),
			},
		},
	}

	_, err := verifier.VerifyHeader(event)
	assert.Nil(t, err)
}
