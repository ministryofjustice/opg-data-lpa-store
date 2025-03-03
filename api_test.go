package main_test

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	authorUID = "34"
)

var (
	baseURL      = cmp.Or(os.Getenv("URL"), "http://localhost:9000")
	jwtSecretKey = cmp.Or(os.Getenv("JWT_SECRET_KEY"), "mysupersecrettestkeythatis128bits")
)

func TestJWTRequired(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping api test")
		return
	}

	t.Run("Put", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut,
			fmt.Sprintf("%s/lpas/%s", baseURL, makeLpaUID()),
			strings.NewReader(`{"version":"1"}`))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("X-Jwt-Authorization", "Bearer "+makeJwt("bad", authorUID))

		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Post", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost,
			fmt.Sprintf("%s/lpas/%s/updates", baseURL, makeLpaUID()),
			strings.NewReader(`{"type":"BUMP_VERSION","changes":[{"key":"/version","old":"1","new":"2"}]}`))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("X-Jwt-Authorization", "Bearer "+makeJwt("bad", authorUID))

		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Get", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet,
			fmt.Sprintf("%s/lpas/%s", baseURL, makeLpaUID()),
			nil)
		req.Header.Add("X-Jwt-Authorization", "Bearer "+makeJwt("bad", authorUID))

		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping api test")
		return
	}

	testcases := map[string]struct {
		in, out string
	}{
		"full": {
			in:  "docs/example-lpa.json",
			out: "docs/example-lpa.json",
		},
		"defaults": {
			in:  "docs/example-lpa-default-request.json",
			out: "docs/example-lpa-default-response.json",
		},
	}

	for scenario, tc := range testcases {
		t.Run(scenario, func(t *testing.T) {
			lpaUID := makeLpaUID()
			inData, _ := os.ReadFile(tc.in)

			req, _ := http.NewRequest(http.MethodPut,
				fmt.Sprintf("%s/lpas/%s", baseURL, lpaUID),
				bytes.NewReader(inData))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("X-Jwt-Authorization", "Bearer "+makeJwt(jwtSecretKey, authorUID))

			resp, _ := http.DefaultClient.Do(req)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)

			getReq, _ := http.NewRequest(http.MethodGet,
				fmt.Sprintf("%s/lpas/%s", baseURL, lpaUID),
				nil)
			getReq.Header.Add("X-Jwt-Authorization", "Bearer "+makeJwt(jwtSecretKey, authorUID))

			getResp, _ := http.DefaultClient.Do(getReq)
			assert.Equal(t, http.StatusOK, getResp.StatusCode)

			var getJSON map[string]any
			json.NewDecoder(getResp.Body).Decode(&getJSON)

			delete(getJSON, "status")
			delete(getJSON, "uid")
			delete(getJSON, "updatedAt")

			outData, _ := os.ReadFile(tc.out)

			getBody, _ := json.Marshal(getJSON)
			assert.JSONEq(t, string(outData), string(getBody))
		})
	}
}

func TestCreateWithMissingFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping api test")
		return
	}

	lpaUID := makeLpaUID()

	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/lpas/%s", baseURL, lpaUID),
		strings.NewReader(`{"version":"2"}`))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Jwt-Authorization", "Bearer "+makeJwt(jwtSecretKey, authorUID))

	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping api test")
		return
	}

	uids := []string{
		doCreateExample(t),
		doCreateExample(t),
		doCreateExample(t),
	}

	getReq, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/lpas", baseURL),
		strings.NewReader(fmt.Sprintf(`{"uids":["%s"]}`, strings.Join(uids, `","`))))
	getReq.Header.Add("X-Jwt-Authorization", "Bearer "+makeJwt(jwtSecretKey, authorUID))

	getResp, _ := http.DefaultClient.Do(getReq)
	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	var getJSON struct {
		Lpas []map[string]any `json:"lpas"`
	}
	json.NewDecoder(getResp.Body).Decode(&getJSON)

	assert.Len(t, getJSON.Lpas, 3)
	assert.Contains(t, uids, getJSON.Lpas[0]["uid"])
	assert.Contains(t, uids, getJSON.Lpas[1]["uid"])
	assert.Contains(t, uids, getJSON.Lpas[2]["uid"])
}

func TestUpdateToStatutoryWaitingPeriod(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping api test")
		return
	}

	steps := []struct {
		name string
		path string
	}{
		{name: "CertificateProviderSign", path: "docs/certificate-provider-sign.json"},
		{name: "AttorneySign", path: "docs/attorney-sign.json"},
		{name: "TrustCorporationSign", path: "docs/trust-corporation-sign.json"},
		{name: "DonorConfirmIdentity", path: "docs/donor-confirm-identity.json"},
		{name: "CertificateProviderConfirmIdentity", path: "docs/certificate-provider-confirm-identity.json"},
		{name: "StatutoryWaitingPeriod", path: "docs/statutory-waiting-period.json"},
	}

	lpaUID := doCreateExample(t)

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			data, _ := os.ReadFile(step.path)

			req, _ := http.NewRequest(http.MethodPost,
				fmt.Sprintf("%s/lpas/%s/updates", baseURL, lpaUID),
				bytes.NewReader(data))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("X-Jwt-Authorization", "Bearer "+makeJwt(jwtSecretKey, authorUID))

			resp, _ := http.DefaultClient.Do(req)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		})
	}
}

func TestCertificateProviderOptOut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping api test")
		return
	}

	lpaUID := doCreateExample(t)

	data, _ := os.ReadFile("docs/certificate-provider-opt-out.json")

	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/lpas/%s/updates", baseURL, lpaUID),
		bytes.NewReader(data))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Jwt-Authorization", "Bearer "+makeJwt(jwtSecretKey, authorUID))

	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestAttorneyOptOut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping api test")
		return
	}

	lpaUID := doCreateExample(t)

	data, _ := os.ReadFile("docs/attorney-opt-out.json")

	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/lpas/%s/updates", baseURL, lpaUID),
		bytes.NewReader(data))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Jwt-Authorization", "Bearer "+makeJwt(jwtSecretKey, "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"))

	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestDonorWithdrawLpa(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping api test")
		return
	}

	lpaUID := doCreateExample(t)

	data, _ := os.ReadFile("docs/donor-withdraw-lpa.json")

	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/lpas/%s/updates", baseURL, lpaUID),
		bytes.NewReader(data))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Jwt-Authorization", "Bearer "+makeJwt(jwtSecretKey, authorUID))

	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func doCreateExample(t *testing.T) string {
	lpaUID := makeLpaUID()
	exampleLpa, _ := os.ReadFile("docs/example-lpa.json")

	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/lpas/%s", baseURL, lpaUID),
		bytes.NewReader(exampleLpa))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Jwt-Authorization", "Bearer "+makeJwt(jwtSecretKey, authorUID))

	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	return lpaUID
}

func makeLpaUID() string {
	return "M-" + strings.ToUpper(uuid.NewString()[9:23])
}

func makeJwt(secretKey, uid string) string {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Add(time.Hour * -24).Unix(),
		"iss": "opg.poas.sirius",
		"sub": "urn:opg:poas:sirius:users:" + uid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))

	if err != nil {
		panic(err)
	}

	return tokenString
}
