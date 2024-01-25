package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

var LPAPath = regexp.MustCompile("^/lpas/(M(?:-[0-9A-Z]{4}){3})$")
var UpdatePath = regexp.MustCompile("^/lpas/(M(?:-[0-9A-Z]{4}){3})/updates$")
var uidMap = map[string]string{}

func delegateHandler(w http.ResponseWriter, r *http.Request) {
	lambdaName := ""
	uid := ""

	if r.URL.Path == "/_pact_state" {
		err := handlePactState(r)
		if err != nil {
			log.Printf("Error setting up state: %s", err.Error())
			http.Error(w, err.Error(), 500)
		} else {
			w.WriteHeader(200)
		}

		return
	}

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

	if newUID, ok := uidMap[uid]; ok {
		uid = newUID
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
		log.Fatal(err)
	}

	encodedRespBody, _ := io.ReadAll(resp.Body)

	var respBody events.APIGatewayProxyResponse
	_ = json.Unmarshal(encodedRespBody, &respBody)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(respBody.StatusCode)
	_, err = w.Write([]byte(respBody.Body))

	if err != nil {
		log.Fatal(err)
	}
}

func handlePactState(r *http.Request) error {
	var state struct {
		State string `json:"state"`
	}

	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		return err
	}

	re := regexp.MustCompile(`^An LPA with UID (M-[A-Z0-9-]+) exists$`)
	if matches := re.FindStringSubmatch(state.State); len(matches) > 0 {
		oldUID := matches[1]
		newUID := randomUID()
		uidMap[oldUID] = newUID

		url := fmt.Sprintf("http://localhost:8080/lpas/%s", oldUID)
		body := `{
			"lpaType": "personal-welfare",
			"donor": {
				"firstNames": "Homer",
				"lastName": "Zoller",
				"dateOfBirth": "1960-04-06",
				"address": {
					"line1": "79 Bury Rd",
					"town": "Hampton Lovett",
					"postcode": "WR9 2PF",
					"country": "GB"
				}
			},
			"attorneys": [
				{
					"firstNames": "Jake",
					"lastName": "Vallar",
					"dateOfBirth": "2001-01-17",
					"status": "active",
					"address": {
						"line1": "71 South Western Terrace",
						"town": "Milton",
						"country": "AU"
					}
				}
			],
			"certificateProvider": {
				"firstNames": "Some",
				"lastName": "Provider",
				"email": "some@example.com",
				"channel": "online",
				"address": {
					"line1": "71 South Western Terrace",
					"town": "Milton",
					"country": "AU"
				}
			},
			"lifeSustainingTreatmentOption": "option-a",
			"signedAt": "2000-01-02T12:13:14Z"
		}`

		req, err := http.NewRequest("PUT", url, strings.NewReader(body))
		if err != nil {
			return err
		}

		req.Header = r.Header.Clone()

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode >= 400 {
			return fmt.Errorf("request failed with status code %d", resp.StatusCode)
		}
	}

	return nil
}

func randomUID() string {
	chunk := func() string {
		bytes := make([]byte, 4)
		_, err := rand.Read(bytes)
		if err != nil {
			panic(err)
		}
		for i, b := range bytes {
			bytes[i] = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"[b%byte(36)]
		}
		return string(bytes)
	}

	return fmt.Sprintf("M-%s-%s-%s", chunk(), chunk(), chunk())
}

func main() {
	http.HandleFunc("/", delegateHandler)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           nil,
		ReadHeaderTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

	log.Println("running on port 8080")
}
