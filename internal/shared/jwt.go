package shared

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt/v5"
)

const (
	sirius string = "opg.poas.sirius"
	mrlpa         = "opg.poas.makeregister"
)

var validIssuers []string = []string{
	sirius,
	mrlpa,
}

type lpaStoreClaims struct {
	jwt.RegisteredClaims
}

// note that default validation for RegisteredClaims checks exp is in the future
func (l lpaStoreClaims) Validate() error {
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

	if iss == sirius {
		emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
		if !emailRegex.MatchString(sub) {
			return errors.New("Subject is not a valid email")
		}
	}

	if iss == mrlpa {
		uidRegex := regexp.MustCompile("^.+$")
		if !uidRegex.MatchString(sub) {
			return errors.New("Subject is not a valid UID")
		}
	}

	return nil
}

type JWTVerifier struct {
	secretKey []byte
}

func NewJWTVerifier() JWTVerifier {
	return JWTVerifier{
		secretKey: []byte(os.Getenv("JWT_SECRET_KEY")),
	}
}

var bearerRegexp = regexp.MustCompile("^Bearer[ ]+")

// verify JWT from event header
// returns true if verified, false otherwise
func (v JWTVerifier) VerifyHeader(event events.APIGatewayProxyRequest) bool {
	jwtHeaders := GetEventHeader("X-Jwt-Authorization", event)

	if len(jwtHeaders) < 1 {
		return false
	}

	tokenStr := bearerRegexp.ReplaceAllString(jwtHeaders[0], "")
	if v.verifyToken(tokenStr) != nil {
		return false
	}

	return true
}

// tokenStr is the JWT token, minus any "Bearer: " prefix
func (v JWTVerifier) verifyToken(tokenStr string) error {
	lsc := lpaStoreClaims{}

	parsedToken, err := jwt.ParseWithClaims(tokenStr, &lsc, func(token *jwt.Token) (interface{}, error) {
		return v.secretKey, nil
	})

	if err != nil {
		return err
	}

	if !parsedToken.Valid {
		return fmt.Errorf("Invalid JWT")
	}

	return nil
}
