package shared

import (
	"testing"
	"time"

    "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

var secretKey = []byte("secret")

func createToken(claims jwt.MapClaims) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    tokenString, err := token.SignedString(secretKey)

    if err != nil {
    	return "", err
    }

 	return tokenString, nil
}

func TestVerifyEmptyJwt(t *testing.T) {
	err := VerifyToken(secretKey, "")
	assert.NotNil(t, err)
}

func TestVerifyExpInPast(t *testing.T) {
	token, _ := createToken(jwt.MapClaims{
        "exp": time.Now().Add(time.Hour * -24).Unix(),
        "iat": time.Now().Add(time.Hour * -24).Unix(),
        "iss": "opg.poas.makeregister",
    })

	err := VerifyToken(secretKey, token)

	assert.NotNil(t, err)
	if err != nil {
		assert.Containsf(t, err.Error(), "token is expired", "")
	}
}

func TestVerifyIatInFuture(t *testing.T) {
	token, _ := createToken(jwt.MapClaims{
        "exp": time.Now().Add(time.Hour * 24).Unix(),
        "iat": time.Now().Add(time.Hour * 24).Unix(),
        "iss": "opg.poas.sirius",
    })

	err := VerifyToken(secretKey, token)

	assert.NotNil(t, err)
	if err != nil {
		assert.Containsf(t, err.Error(), "IssuedAt must not be in the future", "")
	}
}

func TestVerifyIssuer(t *testing.T) {
	token, _ := createToken(jwt.MapClaims{
        "exp": time.Now().Add(time.Hour * 24).Unix(),
        "iat": time.Now().Add(time.Hour * -24).Unix(),
        "iss": "daadsdaadsadsads",
    })

	err := VerifyToken(secretKey, token)

	assert.NotNil(t, err)
	if err != nil {
		assert.Containsf(t, err.Error(), "Invalid Issuer", "")
	}
}

func TestVerifyGoodJwt(t *testing.T) {
	token, _ := createToken(jwt.MapClaims{
        "exp": time.Now().Add(time.Hour * 24).Unix(),
        "iat": time.Now().Add(time.Hour * -24).Unix(),
        "iss": "opg.poas.sirius",
    })

	err := VerifyToken(secretKey, token)

	assert.Nil(t, err)
}
