package shared

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var validIssuers []string = []string{
	"opg.poas.sirius",
	"opg.poas.makeregister",
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

    return nil
}

func VerifyToken(secretKey []byte, tokenStr string) error {
	lsc := lpaStoreClaims{}

 	parsedToken, err := jwt.ParseWithClaims(tokenStr, &lsc, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
   	})

    if err != nil {
    	return err
   	}

   	if !parsedToken.Valid {
    	return fmt.Errorf("invalid token")
   	}

   	return nil
}
