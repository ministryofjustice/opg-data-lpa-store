package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/aws/aws-sdk-go/awstesting/unit"
)

func main() {
	creds := unit.Session.Config.Credentials
	signer := v4.NewSigner(creds)

	method := os.Args[1]
	host := os.Args[2]
	body := strings.NewReader(os.Args[3])

	req, err := http.NewRequest(method, host, body)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-type", "application/json")

	signer.Sign(req, body, "lambda", "eu-west-1", time.Now())

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode >= 400 {
		log.Printf("Response code %d", resp.StatusCode)
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, resp.Body)
		log.Printf("error response: %s", buf.String())

		panic(fmt.Sprintf("invalid status code %d", resp.StatusCode))
	}
}
