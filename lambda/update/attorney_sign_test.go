package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestAttorneySignApply(t *testing.T) {
	attorneyIndex := 1
	lpa := &shared.Lpa{
		LpaInit: shared.LpaInit{
			Attorneys: []shared.Attorney{{}, {}},
		},
	}
	a := AttorneySign{
		Index:                     &attorneyIndex,
		Mobile:                    "0777",
		SignedAt:                  time.Now(),
		ContactLanguagePreference: shared.LangCy,
	}

	errors := a.Apply(lpa)
	assert.Empty(t, errors)
	assert.Equal(t, a.Mobile, lpa.Attorneys[attorneyIndex].Mobile)
	assert.Equal(t, a.SignedAt, lpa.Attorneys[attorneyIndex].SignedAt)
	assert.Equal(t, a.ContactLanguagePreference, lpa.Attorneys[attorneyIndex].ContactLanguagePreference)
}

func TestAttorneySignApplyWhenAlreadySigned(t *testing.T) {
	attorneyIndex := 0
	lpa := &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{{SignedAt: time.Now()}}}}
	a := AttorneySign{
		Index: &attorneyIndex,
	}

	errors := a.Apply(lpa)
	assert.Equal(t, errors, []shared.FieldError{{Source: "/type", Detail: "attorney cannot sign again"}})
}

func TestValidateUpdateAttorneySign(t *testing.T) {
	jsonNull := json.RawMessage("null")

	testcases := map[string]struct {
		update shared.Update
		errors []shared.FieldError
	}{
		"valid": {
			update: shared.Update{
				Type: "ATTORNEY_SIGN",
				Changes: []shared.Change{
					{
						Key: "/attorneys/1/mobile",
						New: json.RawMessage(`"07777"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/1/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/1/contactLanguagePreference",
						New: json.RawMessage(`"cy"`),
						Old: jsonNull,
					},
				},
			},
		},
		"missing all": {
			update: shared.Update{Type: "ATTORNEY_SIGN"},
			errors: []shared.FieldError{
				{Source: "/changes", Detail: "missing /attorneys/..."},
			},
		},
		"extra fields": {
			update: shared.Update{
				Type: "ATTORNEY_SIGN",
				Changes: []shared.Change{
					{
						Key: "/attorneys/1/mobile",
						New: json.RawMessage(`"0777"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/1/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/1/contactLanguagePreference",
						Old: json.RawMessage(`"` + shared.LangEn + `"`),
						New: json.RawMessage(`"` + shared.LangCy + `"`),
					},
					{
						Key: "/donor/firstNames",
						New: json.RawMessage(`"John"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/1/firstNames",
						New: json.RawMessage(`"John"`),
						Old: jsonNull,
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/2/old", Detail: "must not be provided"},
				{Source: "/changes/3", Detail: "unexpected change provided"},
				{Source: "/changes/4", Detail: "unexpected change provided"},
			},
		},
		"invalid contact language": {
			update: shared.Update{
				Type: "ATTORNEY_SIGN",
				Changes: []shared.Change{
					{
						Key: "/attorneys/1/mobile",
						New: json.RawMessage(`"07777"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/1/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/1/contactLanguagePreference",
						New: json.RawMessage(`"xy"`),
						Old: jsonNull,
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/2/new", Detail: "invalid value"},
			},
		},
		"multiple attorneys": {
			update: shared.Update{
				Type: "ATTORNEY_SIGN",
				Changes: []shared.Change{
					{
						Key: "/attorneys/0/mobile",
						New: json.RawMessage(`"0777"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/1/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/0/contactLanguagePreference",
						New: json.RawMessage(`"` + shared.LangCy + `"`),
						Old: jsonNull,
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/1/key", Detail: "index out of range"},
				{Source: "/changes", Detail: "missing /attorneys/0/signedAt"},
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
