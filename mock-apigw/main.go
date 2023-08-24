package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

func delegateHandler(w http.ResponseWriter, r *http.Request) {
	lambdaName := ""

	if r.URL.Path == "/create" && r.Method == http.MethodPost {
		lambdaName = "create"
	}

	if lambdaName == "" {
		http.Error(w, fmt.Sprintf("couldn't match URL: %s", html.EscapeString(r.URL.Path)), http.StatusNotFound)
		return
	}

	url := fmt.Sprintf("http://lambda-%s:8080/2015-03-31/functions/function/invocations", lambdaName)

	reqBody := new(strings.Builder)
	_, _ = io.Copy(reqBody, r.Body)

	body := events.APIGatewayProxyRequest{
		Body:              reqBody.String(),
		HTTPMethod:        r.Method,
		MultiValueHeaders: r.Header,
	}

	encodedBody, _ := json.Marshal(body)

	proxyReq, err := http.NewRequest("POST", url, io.NopCloser(strings.NewReader(string(encodedBody))))
	if err != nil {
		log.Printf("error: couldn't create proxy request")
	}

	client := &http.Client{}
	resp, err := client.Do(proxyReq)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encodedRespBody, _ := io.ReadAll(resp.Body)

	var respBody events.APIGatewayProxyResponse
	_ = json.Unmarshal(encodedRespBody, &respBody)

	w.WriteHeader(respBody.StatusCode)
	w.Write([]byte(respBody.Body))
}

func main() {
	http.HandleFunc("/", delegateHandler)

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
