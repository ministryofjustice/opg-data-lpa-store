package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

func main() {
	sess := session.Must(session.NewSession())
	signer := v4.NewSigner(sess.Config.Credentials)

	expectedStatusCode := flag.Int("expectedStatus", 200, "Expected response status code")
	flag.Parse()

	args := flag.Args()
	method := args[0]

	uid := "M-AL9A-7EY3-075D"
	url := strings.Replace(args[1], "{{UID}}", uid, -1)

	body := strings.NewReader(args[2])

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-type", "application/json")

	_, err = signer.Sign(req, body, "execute-api", "eu-west-1", time.Now())
	if err != nil {
		panic(err)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	buf := new(strings.Builder)
	_, _ = io.Copy(buf, resp.Body)

	log.Printf("*******************")

	if resp.StatusCode != *expectedStatusCode {
		log.Printf("! TEST FAILED - %s to %s", method, url)
		log.Printf("invalid status code %d; expected: %d", resp.StatusCode, *expectedStatusCode)
		log.Printf("error response: %s", buf.String())

	} else {
		log.Printf("Test passed - %d: %s", resp.StatusCode, buf.String())
	}
}
