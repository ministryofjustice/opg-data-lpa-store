package main

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	testcases := map[string]struct {
		certificateProvider CertificateProvider
		errors              []shared.FieldError
	}{
		"valid": {
			certificateProvider: CertificateProvider{
				Address: shared.Address{
					Line1:   "123 Main St",
					Town:    "Homeland",
					Country: "GB",
				},
				SignedAt:                  time.Now(),
				ContactLanguagePreference: shared.LangCy,
			},
		},
		"missing all": {
			errors: []shared.FieldError{
				{Source: "/signedAt", Detail: "field is required"},
				{Source: "/contactLanguagePreference", Detail: "field is required"},
			},
		},
		"bad address": {
			certificateProvider: CertificateProvider{
				Address: shared.Address{
					Line2: "x",
				},
			},
			errors: []shared.FieldError{
				{Source: "/address/line1", Detail: "field is required"},
				{Source: "/address/town", Detail: "field is required"},
				{Source: "/address/country", Detail: "field is required"},
				{Source: "/address/country", Detail: "must be a valid ISO-3166-1 country code"},
				{Source: "/signedAt", Detail: "field is required"},
				{Source: "/contactLanguagePreference", Detail: "field is required"},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.ElementsMatch(t, tc.errors, Validate(tc.certificateProvider))
		})
	}
}
