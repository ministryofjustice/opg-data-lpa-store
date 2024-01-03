package main

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestValidateUpdate(t *testing.T) {
	testcases := map[string]struct {
		update shared.Update
		errors []shared.FieldError
	}{
		"CERTIFICATE_PROVIDER_SIGN/valid": {
			update: shared.Update{
				Type: "CERTIFICATE_PROVIDER_SIGN",
				Changes: []shared.Change{
					{
						Key: "/certificateProvider/address/line1",
						New: "123 Main St",
					},
					{
						Key: "/certificateProvider/address/town",
						New: "Homeland",
					},
					{
						Key: "/certificateProvider/address/country",
						New: "GB",
					},
					{
						Key: "/certificateProvider/signedAt",
						New: time.Now(),
					},
					{
						Key: "/certificateProvider/contactLanguagePreference",
						New: shared.LangCy,
					},
				},
			},
		},
		"CERTIFICATE_PROVIDER_SIGN/missing all": {
			update: shared.Update{Type: "CERTIFICATE_PROVIDER_SIGN"},
			errors: []shared.FieldError{
				{Source: "/changes", Detail: "missing /certificateProvider/signedAt"},
				{Source: "/changes", Detail: "missing /certificateProvider/contactLanguagePreference"},
			},
		},
		"CERTIFICATE_PROVIDER_SIGN/bad address": {
			update: shared.Update{
				Type: "CERTIFICATE_PROVIDER_SIGN",
				Changes: []shared.Change{
					{
						Key: "/certificateProvider/address/country",
						New: "x",
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes", Detail: "missing /certificateProvider/address/line1"},
				{Source: "/changes", Detail: "missing /certificateProvider/address/town"},
				{Source: "/changes/0/new", Detail: "must be a valid ISO-3166-1 country code"},
				{Source: "/changes", Detail: "missing /certificateProvider/signedAt"},
				{Source: "/changes", Detail: "missing /certificateProvider/contactLanguagePreference"},
			},
		},
		"CERTIFICATE_PROVIDER_SIGN/extra fields": {
			update: shared.Update{
				Type: "CERTIFICATE_PROVIDER_SIGN",
				Changes: []shared.Change{
					{
						Key: "/certificateProvider/signedAt",
						New: time.Now(),
					},
					{
						Key: "/certificateProvider/contactLanguagePreference",
						Old: shared.LangEn,
						New: shared.LangCy,
					},
					{
						Key: "/donor/firstNames",
						New: "John",
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/1/old", Detail: "field must not be provided"},
				{Source: "/changes/2", Detail: "change not allowed for type"},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.ElementsMatch(t, tc.errors, validateUpdate(tc.update))
		})
	}
}
