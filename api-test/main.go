package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ./api-test/tester UID -> generate a UID
// ./api-test/tester JWT -> generate a JWT
// JWT_SECRET_KEY=mysupersecrettestkeythatis128bits ./api-test/tester -expectedStatus=200 REQUEST <METHOD> <URL> <REQUEST BODY>
//
//	-> make a test request with a JWT generated using secret "mysupersecrettestkeythatis128bits" and expected status 200
//
// note that the jwtSecret sends a boilerplate JWT for now with valid iat, exp, iss and sub fields
func main() {
	ctx := context.Background()
	expectedStatusCode := flag.Int("expectedStatus", 200, "Expected response status code")
	authorUID := flag.String("authorUID", "34", "Set the UID of the author in the header")
	writeBody := flag.Bool("write", false, "Write the response body to STDOUT")
	flag.Parse()
	args := flag.Args()

	jwtSecret := os.Getenv("JWT_SECRET_KEY")

	switch args[0] {
	case "UID":
		fmt.Print("M-" + strings.ToUpper(uuid.NewString()[9:23]))
		os.Exit(0)
	case "JWT":
		fmt.Print(makeJwt([]byte(jwtSecret), authorUID))
		os.Exit(0)
	case "REQUEST":
		// continue

	default:
		panic("Unrecognised command")
	}

	method := args[1]
	url := args[2]

	var body io.ReadSeeker
	if method != http.MethodGet {
		body = strings.NewReader(args[3])
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}

	if body != nil {
		req.Header.Add("Content-type", "application/json")
	}

	if jwtSecret != "" {
		tokenString := makeJwt([]byte(jwtSecret), authorUID)

		req.Header.Add("X-Jwt-Authorization", fmt.Sprintf("Bearer %s", tokenString))
	}

	if !strings.HasPrefix(url, "http://localhost") {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			panic(err)
		}

		signer := v4.NewSigner()

		credentials, err := cfg.Credentials.Retrieve(ctx)
		if err != nil {
			panic(err)
		}

		hash := sha256.New()
		if body != nil {
			if _, err := io.Copy(hash, body); err != nil {
				panic(err)
			}
			_, _ = body.Seek(0, 0)
		}

		encodedBody := hex.EncodeToString(hash.Sum(nil))

		if err := signer.SignHTTP(ctx, credentials, req, encodedBody, "execute-api", cfg.Region, time.Now()); err != nil {
			panic(err)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, resp.Body)

	log.Printf("*******************")

	if *writeBody {
		fmt.Print(buf.String())
	}

	if resp.StatusCode != *expectedStatusCode {
		log.Printf("! TEST FAILED - %s to %s", method, url)
		log.Printf("invalid status code %d; expected: %d", resp.StatusCode, *expectedStatusCode)
		log.Printf("error response: %s", buf.String())

		os.Exit(1)
	} else {
		log.Print(resp.Header)
		log.Printf("Test passed - %s to %s - %d: %s", method, url, resp.StatusCode, buf.String())
	}
}

func makeJwt(secretKey []byte, uid *string) string {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Add(time.Hour * -24).Unix(),
		"iss": "opg.poas.sirius",
		"sub": "urn:opg:poas:sirius:users:" + *uid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		panic(err)
	}

	return tokenString
}
