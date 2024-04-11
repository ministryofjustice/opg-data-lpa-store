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
	assert.Equal(t, c.SignedAt, *lpa.CertificateProvider.SignedAt)
	assert.Equal(t, c.ContactLanguagePreference, lpa.CertificateProvider.ContactLanguagePreference)
}

func TestCertificateProviderSignApplyWhenAlreadySigned(t *testing.T) {
	signedAt := time.Now()
	lpa := &shared.Lpa{LpaInit: shared.LpaInit{CertificateProvider: shared.CertificateProvider{SignedAt: &signedAt}}}
	c := CertificateProviderSign{}

	errors := c.Apply(lpa)
	assert.Equal(t, errors, []shared.FieldError{{Source: "/type", Detail: "certificate provider cannot sign again"}})
}

func TestValidateUpdateCertificateProviderSign(t *testing.T) {
	jsonNull := json.RawMessage("null")

	testcases := map[string]struct {
		update shared.Update
		errors []shared.FieldError
	}{
		"valid": {
			update: shared.Update{
				Type: "CERTIFICATE_PROVIDER_SIGN",
				Changes: []shared.Change{
					{
						Key: "/certificateProvider/address/line1",
						New: json.RawMessage(`"123 Main St"`),
						Old: jsonNull,
					},
					{
						Key: "/certificateProvider/address/town",
						New: json.RawMessage(`"Homeland"`),
						Old: jsonNull,
					},
					{
						Key: "/certificateProvider/address/country",
						New: json.RawMessage(`"GB"`),
						Old: jsonNull,
					},
					{
						Key: "/certificateProvider/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/certificateProvider/contactLanguagePreference",
						New: json.RawMessage(`"cy"`),
						Old: jsonNull,
					},
					{
						Key: "/certificateProvider/email",
						New: json.RawMessage(`"b@example.com"`),
						Old: json.RawMessage(`"a@example.com"`),
					},
				},
			},
		},
		"missing all": {
			update: shared.Update{Type: "CERTIFICATE_PROVIDER_SIGN"},
			errors: []shared.FieldError{
				{Source: "/changes", Detail: "missing /certificateProvider/signedAt"},
				{Source: "/changes", Detail: "missing /certificateProvider/contactLanguagePreference"},
			},
		},
		"bad address": {
			update: shared.Update{
				Type: "CERTIFICATE_PROVIDER_SIGN",
				Changes: []shared.Change{
					{
						Key: "/certificateProvider/address/line3",
						New: json.RawMessage("1"),
						Old: jsonNull,
					},
					{
						Key: "/certificateProvider/address/country",
						New: json.RawMessage(`"x"`),
						Old: jsonNull,
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0/new", Detail: "unexpected type"},
				{Source: "/changes", Detail: "missing /certificateProvider/address/line1"},
				{Source: "/changes", Detail: "missing /certificateProvider/address/town"},
				{Source: "/changes/1/new", Detail: "must be a valid ISO-3166-1 country code"},
				{Source: "/changes", Detail: "missing /certificateProvider/signedAt"},
				{Source: "/changes", Detail: "missing /certificateProvider/contactLanguagePreference"},
			},
		},
		"extra fields": {
			update: shared.Update{
				Type: "CERTIFICATE_PROVIDER_SIGN",
				Changes: []shared.Change{
					{
						Key: "/certificateProvider/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/certificateProvider/contactLanguagePreference",
						Old: json.RawMessage(`"` + shared.LangEn + `"`),
						New: json.RawMessage(`"` + shared.LangCy + `"`),
					},
					{
						Key: "/donor/firstNames",
						New: json.RawMessage(`"John"`),
						Old: jsonNull,
					},
					{
						Key: "/certificateProvider/email",
						New: json.RawMessage(`"a@example.com"`),
						Old: jsonNull,
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/1/old", Detail: "must be null"},
				{Source: "/changes/2", Detail: "unexpected change provided"},
			},
		},
		"invalid contact language": {
			update: shared.Update{
				Type: "CERTIFICATE_PROVIDER_SIGN",
				Changes: []shared.Change{
					{
						Key: "/certificateProvider/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/certificateProvider/contactLanguagePreference",
						New: json.RawMessage(`"xy"`),
						Old: jsonNull,
					},
					{
						Key: "/certificateProvider/email",
						New: json.RawMessage(`"a@example.com"`),
						Old: jsonNull,
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
