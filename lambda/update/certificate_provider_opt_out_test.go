package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestCertificateProviderOptOutApply(t *testing.T) {
	lpa := &shared.Lpa{Status: shared.LpaStatusInProgress}
	c := CertificateProviderOptOut{}

	errors := c.Apply(lpa)

	assert.Empty(t, errors)
	assert.Equal(t, shared.LpaStatusCannotRegister, lpa.Status)
}

func TestCertificateProviderOptOutApplyWhenCertificateProvided(t *testing.T) {
	now := time.Now()
	certificateProvider := shared.CertificateProvider{Email: "a@example", SignedAt: &now}
	lpa := &shared.Lpa{LpaInit: shared.LpaInit{
		CertificateProvider: certificateProvider},
	}

	errors := CertificateProviderOptOut{}.Apply(lpa)

	assert.Equal(t, errors, []shared.FieldError{{Source: "/type", Detail: "certificate provider cannot opt out after providing certificate"}})
	assert.Equal(t, certificateProvider, lpa.CertificateProvider)
}

func TestValidateUpdateCertificateProviderOptOut(t *testing.T) {
	testcases := map[string]struct {
		update shared.Update
		lpa    *shared.Lpa
		errors []shared.FieldError
	}{
		"valid": {
			update: shared.Update{
				Type:    "CERTIFICATE_PROVIDER_OPT_OUT",
				Changes: []shared.Change{},
			},
		},
		"with changes": {
			update: shared.Update{
				Type: "CERTIFICATE_PROVIDER_OPT_OUT",
				Changes: []shared.Change{
					{
						Key: "/something/someValue",
						New: json.RawMessage(`"not expected"`),
						Old: jsonNull,
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes", Detail: "expected empty"},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			_, errors := validateUpdate(tc.update, &shared.Lpa{})
			assert.ElementsMatch(t, tc.errors, errors)
		})
	}
}
