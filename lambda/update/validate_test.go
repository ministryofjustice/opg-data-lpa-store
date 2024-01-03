package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestCertificateProviderSignApply(t *testing.T) {
	lpa := &shared.Lpa{}
	c := CertificateProviderSign{
		Address:                   shared.Address{Line1: "line 1"},
		SignedAt:                  time.Now(),
		ContactLanguagePreference: shared.LangCy,
	}

	errors := c.Apply(lpa)
	assert.Empty(t, errors)
	assert.Equal(t, c.Address, lpa.CertificateProvider.Address)
	assert.Equal(t, c.SignedAt, lpa.CertificateProvider.SignedAt)
	assert.Equal(t, c.ContactLanguagePreference, lpa.CertificateProvider.ContactLanguagePreference)
}

func TestCertificateProviderSignApplyWhenAlreadySigned(t *testing.T) {
	lpa := &shared.Lpa{LpaInit: shared.LpaInit{CertificateProvider: shared.CertificateProvider{SignedAt: time.Now()}}}
	c := CertificateProviderSign{}

	errors := c.Apply(lpa)
	assert.Equal(t, errors, []shared.FieldError{{Source: "/type", Detail: "certificate provider cannot sign again"}})
}

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
						New: json.RawMessage(`"123 Main St"`),
					},
					{
						Key: "/certificateProvider/address/town",
						New: json.RawMessage(`"Homeland"`),
					},
					{
						Key: "/certificateProvider/address/country",
						New: json.RawMessage(`"GB"`),
					},
					{
						Key: "/certificateProvider/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
					},
					{
						Key: "/certificateProvider/contactLanguagePreference",
						New: json.RawMessage(`"cy"`),
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
						Key: "/certificateProvider/address/line3",
						New: json.RawMessage("1"),
					},
					{
						Key: "/certificateProvider/address/country",
						New: json.RawMessage(`"x"`),
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0/new", Detail: "must be a string"},
				{Source: "/changes", Detail: "missing /certificateProvider/address/line1"},
				{Source: "/changes", Detail: "missing /certificateProvider/address/town"},
				{Source: "/changes/1/new", Detail: "must be a valid ISO-3166-1 country code"},
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
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
					},
					{
						Key: "/certificateProvider/contactLanguagePreference",
						Old: json.RawMessage(`"` + shared.LangEn + `"`),
						New: json.RawMessage(`"` + shared.LangCy + `"`),
					},
					{
						Key: "/donor/firstNames",
						New: json.RawMessage(`"John"`),
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/1/old", Detail: "field must not be provided"},
				{Source: "/changes/2", Detail: "change not allowed for type"},
			},
		},
		"CERTIFICATE_PROVIDER_SIGN/invalid contact language": {
			update: shared.Update{
				Type: "CERTIFICATE_PROVIDER_SIGN",
				Changes: []shared.Change{
					{
						Key: "/certificateProvider/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
					},
					{
						Key: "/certificateProvider/contactLanguagePreference",
						New: json.RawMessage(`"xy"`),
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/1/new", Detail: "invalid value"},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			_, errors := validateUpdate(tc.update)
			assert.ElementsMatch(t, tc.errors, errors)
		})
	}
}
