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
	now := time.Now()
	yesterday := time.Now().Add(-24 * time.Hour)

	testcases := map[string]struct {
		update shared.Update
		lpa    *shared.Lpa
		errors []shared.FieldError
	}{
		"valid - no previous values": {
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					CertificateProvider: shared.CertificateProvider{},
				},
			},
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
						New: json.RawMessage(`"` + now.Format(time.RFC3339Nano) + `"`),
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
						Old: jsonNull,
					},
					{
						Key: "/certificateProvider/channel",
						New: json.RawMessage(`"online"`),
						Old: jsonNull,
					},
				},
			},
		},
		"valid - with previous values": {
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					CertificateProvider: shared.CertificateProvider{
						Address: shared.Address{
							Line1:    "Line 1",
							Line2:    "Line 2",
							Line3:    "Line 3",
							Town:     "Town",
							Postcode: "ABC 123",
							Country:  "GB",
						},
						SignedAt:                  &yesterday,
						ContactLanguagePreference: shared.LangEn,
						Email:                     "a@example.com",
						Channel:                   shared.ChannelOnline,
					},
				},
			},
			update: shared.Update{
				Type: "CERTIFICATE_PROVIDER_SIGN",
				Changes: []shared.Change{
					{
						Key: "/certificateProvider/address/line1",
						New: json.RawMessage(`"New Line 1"`),
						Old: json.RawMessage(`"Line 1"`),
					},
					{
						Key: "/certificateProvider/address/line2",
						New: json.RawMessage(`"New Line 2"`),
						Old: json.RawMessage(`"Line 2"`),
					},
					{
						Key: "/certificateProvider/address/line3",
						New: json.RawMessage(`"New Line 3"`),
						Old: json.RawMessage(`"Line 3"`),
					},
					{
						Key: "/certificateProvider/address/town",
						New: json.RawMessage(`"New Town"`),
						Old: json.RawMessage(`"Town"`),
					},
					{
						Key: "/certificateProvider/address/country",
						New: json.RawMessage(`"FR"`),
						Old: json.RawMessage(`"GB"`),
					},
					{
						Key: "/certificateProvider/signedAt",
						New: json.RawMessage(`"` + now.Format(time.RFC3339Nano) + `"`),
						Old: json.RawMessage(`"` + yesterday.Format(time.RFC3339Nano) + `"`),
					},
					{
						Key: "/certificateProvider/contactLanguagePreference",
						New: json.RawMessage(`"cy"`),
						Old: json.RawMessage(`"en"`),
					},
					{
						Key: "/certificateProvider/email",
						New: json.RawMessage(`"b@example.com"`),
						Old: json.RawMessage(`"a@example.com"`),
					},
					{
						Key: "/certificateProvider/channel",
						New: json.RawMessage(`"paper"`),
						Old: json.RawMessage(`"online"`),
					},
				},
			},
		},
		"valid - can exclude optional fields": {
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					CertificateProvider: shared.CertificateProvider{},
				},
			},
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
						New: json.RawMessage(`"` + now.Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/certificateProvider/contactLanguagePreference",
						New: json.RawMessage(`"cy"`),
						Old: jsonNull,
					},
				},
			},
		},
		"missing all": {
			lpa:    &shared.Lpa{},
			update: shared.Update{Type: "CERTIFICATE_PROVIDER_SIGN"},
			errors: []shared.FieldError{
				{Source: "/positionChanges", Detail: "missing /certificateProvider/signedAt"},
			},
		},
		"bad address": {
			lpa: &shared.Lpa{},
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
				{Source: "/positionChanges/0/new", Detail: "unexpected type"},
				{Source: "/positionChanges", Detail: "missing /certificateProvider/address/line1"},
				{Source: "/positionChanges", Detail: "missing /certificateProvider/address/town"},
				{Source: "/positionChanges/1/new", Detail: "must be a valid ISO-3166-1 country code"},
				{Source: "/positionChanges", Detail: "missing /certificateProvider/signedAt"},
			},
		},
		"extra fields": {
			lpa: &shared.Lpa{},
			update: shared.Update{
				Type: "CERTIFICATE_PROVIDER_SIGN",
				Changes: []shared.Change{
					{
						Key: "/certificateProvider/signedAt",
						New: json.RawMessage(`"` + now.Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/certificateProvider/contactLanguagePreference",
						New: json.RawMessage(`"` + shared.LangEn + `"`),
						Old: jsonNull,
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
				},
			},
			errors: []shared.FieldError{
				{Source: "/positionChanges/2", Detail: "unexpected change provided"},
			},
		},
		"invalid contact language": {
			lpa: &shared.Lpa{},
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
				},
			},
			errors: []shared.FieldError{
				{Source: "/positionChanges/1/new", Detail: "invalid value"},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			_, errors := validateUpdate(tc.update, tc.lpa)
			assert.ElementsMatch(t, tc.errors, errors)
		})
	}
}
