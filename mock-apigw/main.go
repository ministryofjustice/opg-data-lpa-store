package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

var LPAPath = regexp.MustCompile("^/lpas/(M(?:-[0-9A-Z]{4}){3})$")
var UpdatePath = regexp.MustCompile("^/lpas/(M(?:-[0-9A-Z]{4}){3})/updates$")

func delegateHandler(w http.ResponseWriter, r *http.Request) {
	lambdaName := ""
	uid := ""

	if LPAPath.MatchString(r.URL.Path) && r.Method == http.MethodPut {
		uid = LPAPath.FindStringSubmatch(r.URL.Path)[1]
		lambdaName = "create"
	} else if LPAPath.MatchString(r.URL.Path) && r.Method == http.MethodGet {
		uid = LPAPath.FindStringSubmatch(r.URL.Path)[1]
		lambdaName = "get"
	} else if UpdatePath.MatchString(r.URL.Path) && r.Method == http.MethodPost {
		uid = UpdatePath.FindStringSubmatch(r.URL.Path)[1]
		lambdaName = "update"
	}

	if lambdaName == "" {
		http.Error(w, fmt.Sprintf("couldn't match URL: %s", html.EscapeString(r.URL.Path)), http.StatusNotFound)
		return
	}

	url := fmt.Sprintf("http://lambda-%s:8080/2015-03-31/functions/function/invocations", lambdaName)

	reqBody := new(strings.Builder)
	_, _ = io.Copy(reqBody, r.Body)

	body := events.APIGatewayProxyRequest{
		Body: reqBody.String(),
		Path: r.URL.Path,
		PathParameters: map[string]string{
			"uid": uid,
		},
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
