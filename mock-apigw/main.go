package main

import (
	"bytes"
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
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

var LPAPath = regexp.MustCompile("^/lpas/(M(?:-[0-9A-Z]{4}){3})$")
var UpdatePath = regexp.MustCompile("^/lpas/(M(?:-[0-9A-Z]{4}){3})/updates$")

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
	if match := re.FindStringSubmatch(state.State); len(match) > 0 {
		url := fmt.Sprintf("http://localhost:8080/lpas/%s", match[1])

		lpa := shared.LpaInit{
			LpaType: shared.LpaTypePersonalWelfare,
			Donor: shared.Donor{
				Person: shared.Person{
					FirstNames: "Homer",
					LastName:   "Zoller",
					Address: shared.Address{
						Line1:    "79 Bury Rd",
						Town:     "Hampton Lovett",
						Postcode: "WR9 2PF",
						Country:  "GB",
					},
				},
				DateOfBirth: shared.Date{Time: time.Date(1960, time.April, 6, 0, 0, 0, 0, time.UTC)},
			},
			Attorneys: []shared.Attorney{{
				Person: shared.Person{
					FirstNames: "Jake",
					LastName:   "Vallar",
					Address: shared.Address{
						Line1:   "71 South Western Terrace",
						Town:    "Milton",
						Country: "AU",
					},
				},
				DateOfBirth: shared.Date{Time: time.Date(2001, time.January, 17, 0, 0, 0, 0, time.UTC)},
				Status:      shared.AttorneyStatusActive,
			}},
			CertificateProvider: shared.CertificateProvider{
				Person: shared.Person{
					FirstNames: "Some",
					LastName:   "Provider",
					Address: shared.Address{
						Line1:   "71 South Western Terrace",
						Town:    "Milton",
						Country: "AU",
					},
				},
				Email:   "some@example.com",
				Channel: shared.ChannelOnline,
			},
			LifeSustainingTreatmentOption: shared.LifeSustainingTreatmentOptionA,
			SignedAt:                      time.Date(2000, time.January, 2, 12, 13, 14, 0, time.UTC),
		}

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(lpa); err != nil {
			return err
		}

		req, err := http.NewRequest("PUT", url, &buf)
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
