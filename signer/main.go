package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/google/uuid"
)

// call with UID to generate a UID, or with
// -expectedStatus=200 REQUEST <METHOD> <URL> <REQUEST BODY> to make a test request
func main() {
	expectedStatusCode := flag.Int("expectedStatus", 200, "Expected response status code")
	flag.Parse()

	args := flag.Args()

	// early exit if we're just generating a UID
	if args[0] == "UID" {
		fmt.Print("M-" + strings.ToUpper(uuid.NewString()[9:23]))
		os.Exit(0)
	}

	if args[0] != "REQUEST" {
		panic("Unrecognised command")
	}

	sess := session.Must(session.NewSession())
	signer := v4.NewSigner(sess.Config.Credentials)

	method := args[1]
	url := args[2]
	body := strings.NewReader(args[3])

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
