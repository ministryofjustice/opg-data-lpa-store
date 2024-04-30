package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestTrustCorporationSignApply(t *testing.T) {
	trustCorporationIndex := 1
	lpa := &shared.Lpa{
		LpaInit: shared.LpaInit{
			TrustCorporations: []shared.TrustCorporation{{}, {}},
		},
	}
	a := TrustCorporationSign{
		Index:                     &trustCorporationIndex,
		Mobile:                    "0777",
		Signatories:               [2]shared.Signatory{{SignedAt: time.Now()}},
		ContactLanguagePreference: shared.LangCy,
	}

	errors := a.Apply(lpa)
	assert.Empty(t, errors)
	assert.Equal(t, a.Mobile, lpa.TrustCorporations[trustCorporationIndex].Mobile)
	assert.Equal(t, []shared.Signatory{{SignedAt: a.Signatories[0].SignedAt}}, lpa.TrustCorporations[trustCorporationIndex].Signatories)
	assert.Equal(t, a.ContactLanguagePreference, lpa.TrustCorporations[trustCorporationIndex].ContactLanguagePreference)
}

func TestTrustCorporationSignApplyWhenAlreadySigned(t *testing.T) {
	trustCorporationIndex := 0
	lpa := &shared.Lpa{
		LpaInit: shared.LpaInit{
			TrustCorporations: []shared.TrustCorporation{{
				Signatories: []shared.Signatory{{SignedAt: time.Now()}},
			}},
		},
	}
	a := TrustCorporationSign{
		Index: &trustCorporationIndex,
	}

	errors := a.Apply(lpa)
	assert.Equal(t, errors, []shared.FieldError{{Source: "/type", Detail: "trust corporation cannot sign again"}})
}

func TestValidateUpdateTrustCorporationSign(t *testing.T) {
	testcases := map[string]struct {
		update shared.Update
		errors []shared.FieldError
	}{
		"valid": {
			update: shared.Update{
				Type: "TRUST_CORPORATION_SIGN",
				Changes: []shared.Change{
					{
						Key: "/trustCorporations/1/mobile",
						New: json.RawMessage(`"07777"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/firstNames",
						New: json.RawMessage(`"John"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/lastName",
						New: json.RawMessage(`"Smith"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/professionalTitle",
						New: json.RawMessage(`"Director"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/1/firstNames",
						New: json.RawMessage(`"Jane"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/1/lastName",
						New: json.RawMessage(`"Smith"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/1/professionalTitle",
						New: json.RawMessage(`"Deputy Director"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/1/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/contactLanguagePreference",
						New: json.RawMessage(`"cy"`),
						Old: jsonNull,
					},
				},
			},
		},
		"valid - existing values": {
			update: shared.Update{
				Type: "TRUST_CORPORATION_SIGN",
				Changes: []shared.Change{
					{
						Key: "/trustCorporations/1/mobile",
						New: json.RawMessage(`"07777"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/firstNames",
						New: json.RawMessage(`"John"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/lastName",
						New: json.RawMessage(`"Smith"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/professionalTitle",
						New: json.RawMessage(`"Director"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/contactLanguagePreference",
						New: json.RawMessage(`"cy"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/email",
						New: json.RawMessage(`"b@example.com"`),
						Old: json.RawMessage(`"a@example.com"`),
					},
					{
						Key: "/trustCorporations/1/channel",
						New: json.RawMessage(`"online"`),
						Old: json.RawMessage(`"paper"`),
					},
				},
			},
		},
		"missing all": {
			update: shared.Update{Type: "TRUST_CORPORATION_SIGN"},
			errors: []shared.FieldError{
				{Source: "/changes", Detail: "missing /trustCorporations/..."},
			},
		},
		"extra fields": {
			update: shared.Update{
				Type: "TRUST_CORPORATION_SIGN",
				Changes: []shared.Change{
					{
						Key: "/trustCorporations/1/mobile",
						New: json.RawMessage(`"0777"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/firstNames",
						New: json.RawMessage(`"John"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/lastName",
						New: json.RawMessage(`"Smith"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/professionalTitle",
						New: json.RawMessage(`"Director"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/contactLanguagePreference",
						New: json.RawMessage(`"` + shared.LangCy + `"`),
						Old: jsonNull,
					},
					{
						Key: "/donor/firstNames",
						New: json.RawMessage(`"John"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/firstNames",
						New: json.RawMessage(`"John"`),
						Old: jsonNull,
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/6", Detail: "unexpected change provided"},
				{Source: "/changes/7", Detail: "unexpected change provided"},
			},
		},
		"invalid contact language": {
			update: shared.Update{
				Type: "TRUST_CORPORATION_SIGN",
				Changes: []shared.Change{
					{
						Key: "/trustCorporations/1/mobile",
						New: json.RawMessage(`"07777"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/firstNames",
						New: json.RawMessage(`"John"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/lastName",
						New: json.RawMessage(`"Smith"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/professionalTitle",
						New: json.RawMessage(`"Director"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/contactLanguagePreference",
						New: json.RawMessage(`"xy"`),
						Old: jsonNull,
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/5/new", Detail: "invalid value"},
			},
		},
		"multiple trust corporations": {
			update: shared.Update{
				Type: "TRUST_CORPORATION_SIGN",
				Changes: []shared.Change{
					{
						Key: "/trustCorporations/0/mobile",
						New: json.RawMessage(`"0777"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/1/signatories/0/firstNames",
						New: json.RawMessage(`"John"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/0/signatories/0/lastName",
						New: json.RawMessage(`"Smith"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/0/signatories/0/professionalTitle",
						New: json.RawMessage(`"Director"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/0/signatories/0/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/trustCorporations/0/contactLanguagePreference",
						New: json.RawMessage(`"` + shared.LangCy + `"`),
						Old: jsonNull,
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/1/key", Detail: "index out of range"},
				{Source: "/changes", Detail: "missing /trustCorporations/0/signatories/0/firstNames"},
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
