package main_test

import (
	"bytes"
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
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

	examplePath      = "docs/example-lpa.json"
	exampleImagePath = "docs/example-lpa-images.json"
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
		withAuth(req, "bad", authorUID)

		resp, _ := doRequest(req)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Post", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost,
			fmt.Sprintf("%s/lpas/%s/updates", baseURL, makeLpaUID()),
			strings.NewReader(`{"type":"BUMP_VERSION","changes":[{"key":"/version","old":"1","new":"2"}]}`))
		withAuth(req, "bad", authorUID)

		resp, _ := doRequest(req)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Get", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet,
			fmt.Sprintf("%s/lpas/%s", baseURL, makeLpaUID()),
			nil)
		withAuth(req, "bad", authorUID)

		resp, _ := doRequest(req)
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
			withAuth(req, jwtSecretKey, authorUID)

			resp, _ := doRequest(req)
			if !assert.Equal(t, http.StatusCreated, resp.StatusCode) {
				return
			}

			getReq, _ := http.NewRequest(http.MethodGet,
				fmt.Sprintf("%s/lpas/%s", baseURL, lpaUID),
				nil)
			withAuth(getReq, jwtSecretKey, authorUID)

			getResp, _ := doRequest(getReq)
			if !assert.Equal(t, http.StatusOK, getResp.StatusCode) {
				return
			}

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

func TestCreateWithImages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping api test")
		return
	}

	testcases := map[string]struct {
		urlFormat string
		pathRe    func(string) string
	}{
		"Plain": {
			urlFormat: "%s/lpas/%s",
			pathRe: func(lpaUID string) string {
				return "^" + lpaUID + "/scans/rc_0_my-restrictions.png$"
			},
		},
		"Presigned": {
			urlFormat: "%s/lpas/%s?presign-images",
			pathRe: func(lpaUID string) string {
				hostBucket := "http?://localstack:4566/"
				if !strings.HasPrefix(baseURL, "http://localhost") {
					hostBucket = "https://s3.eu-west-1.amazonaws.com/[a-z0-9\\-]+/"
				}

				return hostBucket + lpaUID + "/scans/rc_0_my-restrictions.png\\?X-Amz-Algorithm=AWS4-HMAC-SHA256&.+$"
			},
		},
	}

	lpaUID := doCreateExample(t, exampleImagePath)

	for scenario, tc := range testcases {
		t.Run(scenario, func(t *testing.T) {
			getReq, _ := http.NewRequest(http.MethodGet,
				fmt.Sprintf(tc.urlFormat, baseURL, lpaUID),
				nil)
			withAuth(getReq, jwtSecretKey, authorUID)

			getResp, _ := doRequest(getReq)
			if !assert.Equal(t, http.StatusOK, getResp.StatusCode) {
				return
			}

			var getJSON map[string]json.RawMessage
			json.NewDecoder(getResp.Body).Decode(&getJSON)

			var restrictionsAndConditionsImages []map[string]string
			json.Unmarshal(getJSON["restrictionsAndConditionsImages"], &restrictionsAndConditionsImages)

			getJSON["channel"] = json.RawMessage(`"online"`)
			delete(getJSON, "status")
			delete(getJSON, "uid")
			delete(getJSON, "updatedAt")
			delete(getJSON, "restrictionsAndConditionsImages")

			outData, _ := os.ReadFile("docs/example-lpa.json")

			getBody, _ := json.Marshal(getJSON)
			assert.JSONEq(t, string(outData), string(getBody))

			assert.Regexp(t, tc.pathRe(lpaUID), restrictionsAndConditionsImages[0]["path"])
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
	withAuth(req, jwtSecretKey, authorUID)

	resp, _ := doRequest(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping api test")
		return
	}

	uids := []string{
		doCreateExample(t, examplePath),
		doCreateExample(t, examplePath),
		doCreateExample(t, examplePath),
	}

	getReq, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/lpas", baseURL),
		strings.NewReader(fmt.Sprintf(`{"uids":["%s"]}`, strings.Join(uids, `","`))))
	withAuth(getReq, jwtSecretKey, authorUID)

	getResp, _ := doRequest(getReq)
	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	var getJSON struct {
		Lpas []struct {
			UID string `json:"uid"`
		} `json:"lpas"`
	}
	json.NewDecoder(getResp.Body).Decode(&getJSON)

	assert.Len(t, getJSON.Lpas, 3)
	assert.Contains(t, uids, getJSON.Lpas[0].UID)
	assert.Contains(t, uids, getJSON.Lpas[1].UID)
	assert.Contains(t, uids, getJSON.Lpas[2].UID)
}

func TestGetListWithImages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping api test")
		return
	}

	testcases := map[string]struct {
		urlFormat string
		pathRe    func(string) string
	}{
		"Plain": {
			urlFormat: "%s/lpas",
			pathRe: func(lpaUID string) string {
				return "^" + lpaUID + "/scans/rc_0_my-restrictions.png$"
			},
		},
		"Presigned": {
			urlFormat: "%s/lpas?presign-images",
			pathRe: func(lpaUID string) string {
				hostBucket := "http://localstack:4566/"
				if !strings.HasPrefix(baseURL, "http://localhost") {
					hostBucket = "https://s3.eu-west-1.amazonaws.com/[a-z0-9\\-]+/"
				}

				return hostBucket + lpaUID + "/scans/rc_0_my-restrictions.png\\?X-Amz-Algorithm=AWS4-HMAC-SHA256&.+$"
			},
		},
	}

	uids := []string{
		doCreateExample(t, exampleImagePath),
		doCreateExample(t, exampleImagePath),
		doCreateExample(t, exampleImagePath),
	}

	for scenario, tc := range testcases {
		t.Run(scenario, func(t *testing.T) {
			getReq, _ := http.NewRequest(http.MethodPost,
				fmt.Sprintf(tc.urlFormat, baseURL),
				strings.NewReader(fmt.Sprintf(`{"uids":["%s"]}`, strings.Join(uids, `","`))))
			withAuth(getReq, jwtSecretKey, authorUID)

			getResp, _ := doRequest(getReq)
			assert.Equal(t, http.StatusOK, getResp.StatusCode)

			var getJSON struct {
				Lpas []struct {
					UID                             string `json:"uid"`
					RestrictionsAndConditionsImages []struct {
						Path string `json:"path"`
					} `json:"restrictionsAndConditionsImages"`
				} `json:"lpas"`
			}
			json.NewDecoder(getResp.Body).Decode(&getJSON)

			assert.Len(t, getJSON.Lpas, 3)
			assert.Contains(t, uids, getJSON.Lpas[0].UID)
			assert.Contains(t, uids, getJSON.Lpas[1].UID)
			assert.Contains(t, uids, getJSON.Lpas[2].UID)

			for _, lpa := range getJSON.Lpas {
				restrictionsAndConditionsImages := lpa.RestrictionsAndConditionsImages

				assert.Regexp(t, tc.pathRe(lpa.UID), restrictionsAndConditionsImages[0].Path)
			}
		})
	}
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

	lpaUID := doCreateExample(t, examplePath)

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			data, _ := os.ReadFile(step.path)

			req, _ := http.NewRequest(http.MethodPost,
				fmt.Sprintf("%s/lpas/%s/updates", baseURL, lpaUID),
				bytes.NewReader(data))
			withAuth(req, jwtSecretKey, authorUID)

			resp, _ := doRequest(req)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		})
	}
}

func TestUpdatesEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping api test")
		return
	}

	testcases := map[string]struct {
		path      string
		authorUID string
	}{
		"CertificateProviderOptOut": {
			path:      "docs/certificate-provider-opt-out.json",
			authorUID: authorUID,
		},
		"AttorneyOptOut": {
			path:      "docs/attorney-opt-out.json",
			authorUID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d",
		},
		"DonorWithdrawLpa": {
			path:      "docs/donor-withdraw-lpa.json",
			authorUID: authorUID,
		},
	}

	for scenario, tc := range testcases {
		t.Run(scenario, func(t *testing.T) {
			lpaUID := doCreateExample(t, examplePath)

			data, _ := os.ReadFile(tc.path)

			req, _ := http.NewRequest(http.MethodPost,
				fmt.Sprintf("%s/lpas/%s/updates", baseURL, lpaUID),
				bytes.NewReader(data))
			withAuth(req, jwtSecretKey, tc.authorUID)

			resp, _ := doRequest(req)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		})
	}
}

func doCreateExample(t *testing.T, examplePath string) string {
	lpaUID := makeLpaUID()
	exampleLpa, _ := os.ReadFile(examplePath)

	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/lpas/%s", baseURL, lpaUID),
		bytes.NewReader(exampleLpa))
	withAuth(req, jwtSecretKey, authorUID)

	resp, _ := doRequest(req)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	return lpaUID
}

func doRequest(req *http.Request) (*http.Response, error) {
	if req.URL.Hostname() == "localhost" {
		return http.DefaultClient.Do(req)
	}

	cfg, err := config.LoadDefaultConfig(req.Context())
	if err != nil {
		return nil, err
	}

	signer := v4.NewSigner()

	credentials, err := cfg.Credentials.Retrieve(req.Context())
	if err != nil {
		return nil, err
	}

	hash := sha256.New()
	if req.Body != nil {
		var reqBody bytes.Buffer

		if _, err := io.Copy(hash, io.TeeReader(req.Body, &reqBody)); err != nil {
			return nil, err
		}

		req.Header.Add("Content-Type", "application/json")
		req.Body = io.NopCloser(&reqBody)
	}

	encodedBody := hex.EncodeToString(hash.Sum(nil))

	if err := signer.SignHTTP(req.Context(), credentials, req, encodedBody, "execute-api", cfg.Region, time.Now()); err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req)
}

func makeLpaUID() string {
	return "M-" + strings.ToUpper(uuid.NewString()[9:23])
}

func withAuth(req *http.Request, jwtSecretKey, authorUID string) {
	req.Header.Add("X-Jwt-Authorization", "Bearer "+makeJwt(jwtSecretKey, authorUID))
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
