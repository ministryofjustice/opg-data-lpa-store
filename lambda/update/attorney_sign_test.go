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
		Channel:                   shared.ChannelOnline,
	}

	errors := a.Apply(lpa)
	assert.Empty(t, errors)
	assert.Equal(t, a.Mobile, lpa.Attorneys[attorneyIndex].Mobile)
	assert.Equal(t, a.SignedAt, *lpa.Attorneys[attorneyIndex].SignedAt)
	assert.Equal(t, a.ContactLanguagePreference, lpa.Attorneys[attorneyIndex].ContactLanguagePreference)
	assert.Equal(t, a.Channel, lpa.Attorneys[attorneyIndex].Channel)
}

func TestAttorneySignApplyWhenAlreadySigned(t *testing.T) {
	attorneyIndex := 0
	signedAt := time.Now()
	lpa := &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{{SignedAt: &signedAt}}}}
	a := AttorneySign{
		Index: &attorneyIndex,
	}

	errors := a.Apply(lpa)
	assert.Equal(t, errors, []shared.FieldError{{Source: "/type", Detail: "attorney cannot sign again"}})
}

func TestValidateUpdateAttorneySign(t *testing.T) {
	now := time.Now()
	yesterday := time.Now()

	testcases := map[string]struct {
		update shared.Update
		lpa    *shared.Lpa
		errors []shared.FieldError
	}{
		"valid - no previous values": {
			update: shared.Update{
				Type: "ATTORNEY_SIGN",
				Changes: []shared.Change{
					{
						Key: "/attorneys/0/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339Nano) + `"`),
						Old: jsonNull,
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{},
			}}},
		},
		"valid - with previous values": {
			update: shared.Update{
				Type: "ATTORNEY_SIGN",
				Changes: []shared.Change{
					{
						Key: "/attorneys/0/mobile",
						New: json.RawMessage(`"07777"`),
						Old: json.RawMessage(`"06666"`),
					},
					{
						Key: "/attorneys/0/signedAt",
						New: json.RawMessage(`"` + now.Format(time.RFC3339Nano) + `"`),
						Old: json.RawMessage(`"` + yesterday.Format(time.RFC3339Nano) + `"`),
					},
					{
						Key: "/attorneys/0/contactLanguagePreference",
						New: json.RawMessage(`"cy"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/0/channel",
						New: json.RawMessage(`"online"`),
						Old: json.RawMessage(`"paper"`),
					},
					{
						Key: "/attorneys/0/email",
						New: json.RawMessage(`"b@example.com"`),
						Old: json.RawMessage(`"a@example.com"`),
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{Channel: shared.ChannelPaper, Email: "a@example.com", Mobile: "06666", SignedAt: &yesterday},
			}}},
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
						Key: "/attorneys/0/mobile",
						New: json.RawMessage(`"0777"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/0/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/0/contactLanguagePreference",
						New: json.RawMessage(`"` + shared.LangCy + `"`),
						Old: jsonNull,
					},
					{
						Key: "/donor/firstNames",
						New: json.RawMessage(`"John"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/0/firstNames",
						New: json.RawMessage(`"John"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/0/email",
						New: json.RawMessage(`"a@example.com"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/0/channel",
						New: json.RawMessage(`"paper"`),
						Old: jsonNull,
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{},
			}}},
			errors: []shared.FieldError{
				{Source: "/changes/3", Detail: "unexpected change provided"},
				{Source: "/changes/4", Detail: "unexpected change provided"},
			},
		},
		"invalid contact language and channel": {
			update: shared.Update{
				Type: "ATTORNEY_SIGN",
				Changes: []shared.Change{
					{
						Key: "/attorneys/0/mobile",
						New: json.RawMessage(`"07777"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/0/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/0/contactLanguagePreference",
						New: json.RawMessage(`"xy"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/0/channel",
						New: json.RawMessage(`"digital"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/0/email",
						New: json.RawMessage(`"b@example.com"`),
						Old: jsonNull,
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{},
			}}},
			errors: []shared.FieldError{
				{Source: "/changes/2/new", Detail: "invalid value"},
				{Source: "/changes/3/new", Detail: "invalid value"},
			},
		},
		"multiple attorneys - multiple attorney changes": {
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
					{
						Key: "/attorneys/0/email",
						New: json.RawMessage(`"a@example.com"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/0/channel",
						New: json.RawMessage(`"online"`),
						Old: jsonNull,
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{}, {},
			}}},
			errors: []shared.FieldError{
				{Source: "/changes/1/key", Detail: "index out of range"},
				{Source: "/changes", Detail: "missing /attorneys/0/signedAt"},
			},
		},
		"multiple attorneys - single attorney change": {
			update: shared.Update{
				Type: "ATTORNEY_SIGN",
				Changes: []shared.Change{
					{
						Key: "/attorneys/1/mobile",
						New: json.RawMessage(`"07777"`),
						Old: json.RawMessage(`"06666"`),
					},
					{
						Key: "/attorneys/1/signedAt",
						New: json.RawMessage(`"` + now.Format(time.RFC3339Nano) + `"`),
						Old: json.RawMessage(`"` + yesterday.Format(time.RFC3339Nano) + `"`),
					},
					{
						Key: "/attorneys/1/contactLanguagePreference",
						New: json.RawMessage(`"cy"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/1/channel",
						New: json.RawMessage(`"online"`),
						Old: json.RawMessage(`"paper"`),
					},
					{
						Key: "/attorneys/1/email",
						New: json.RawMessage(`"b@example.com"`),
						Old: json.RawMessage(`"a@example.com"`),
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{}, {Channel: shared.ChannelPaper, Email: "a@example.com", Mobile: "06666", SignedAt: &yesterday},
			}}},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			_, errors := validateUpdate(tc.update, tc.lpa)
			assert.ElementsMatch(t, tc.errors, errors)
		})
	}
}

func TestValidateUpdateAttorneySignActorUID(t *testing.T) {
	now := time.Now()
	yesterday := time.Now()

	testcases := map[string]struct {
		update shared.Update
		lpa    *shared.Lpa
		errors []shared.FieldError
	}{
		"valid - no previous values": {
			update: shared.Update{
				Type: "ATTORNEY_SIGN",
				Changes: []shared.Change{
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339Nano) + `"`),
						Old: jsonNull,
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{Person: shared.Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"}},
			}}},
		},
		"valid - with previous values": {
			update: shared.Update{
				Type: "ATTORNEY_SIGN",
				Changes: []shared.Change{
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/mobile",
						New: json.RawMessage(`"07777"`),
						Old: json.RawMessage(`"06666"`),
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/signedAt",
						New: json.RawMessage(`"` + now.Format(time.RFC3339Nano) + `"`),
						Old: json.RawMessage(`"` + yesterday.Format(time.RFC3339Nano) + `"`),
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/contactLanguagePreference",
						New: json.RawMessage(`"cy"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/channel",
						New: json.RawMessage(`"online"`),
						Old: json.RawMessage(`"paper"`),
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/email",
						New: json.RawMessage(`"b@example.com"`),
						Old: json.RawMessage(`"a@example.com"`),
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{Person: shared.Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"}, Channel: shared.ChannelPaper, Email: "a@example.com", Mobile: "06666", SignedAt: &yesterday},
			}}},
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
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/mobile",
						New: json.RawMessage(`"0777"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/contactLanguagePreference",
						New: json.RawMessage(`"` + shared.LangCy + `"`),
						Old: jsonNull,
					},
					{
						Key: "/donor/firstNames",
						New: json.RawMessage(`"John"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/firstNames",
						New: json.RawMessage(`"John"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/email",
						New: json.RawMessage(`"a@example.com"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/channel",
						New: json.RawMessage(`"paper"`),
						Old: jsonNull,
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{Person: shared.Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"}},
			}}},
			errors: []shared.FieldError{
				{Source: "/changes/3", Detail: "unexpected change provided"},
				{Source: "/changes/4", Detail: "unexpected change provided"},
			},
		},
		"invalid contact language and channel": {
			update: shared.Update{
				Type: "ATTORNEY_SIGN",
				Changes: []shared.Change{
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/mobile",
						New: json.RawMessage(`"07777"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/contactLanguagePreference",
						New: json.RawMessage(`"xy"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/channel",
						New: json.RawMessage(`"digital"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/email",
						New: json.RawMessage(`"b@example.com"`),
						Old: jsonNull,
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{Person: shared.Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"}},
			}}},
			errors: []shared.FieldError{
				{Source: "/changes/2/new", Detail: "invalid value"},
				{Source: "/changes/3/new", Detail: "invalid value"},
			},
		},
		"multiple attorneys - multiple attorney changes": {
			update: shared.Update{
				Type: "ATTORNEY_SIGN",
				Changes: []shared.Change{
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/mobile",
						New: json.RawMessage(`"0777"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3e/signedAt",
						New: json.RawMessage(`"` + time.Now().Format(time.RFC3339) + `"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/contactLanguagePreference",
						New: json.RawMessage(`"` + shared.LangCy + `"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/email",
						New: json.RawMessage(`"a@example.com"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/channel",
						New: json.RawMessage(`"online"`),
						Old: jsonNull,
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{Person: shared.Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"}},
			}}},
			errors: []shared.FieldError{
				{Source: "/changes/1/key", Detail: "index out of range"},
				{Source: "/changes", Detail: "missing /attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/signedAt"},
			},
		},
		"multiple attorneys - single attorney change": {
			update: shared.Update{
				Type: "ATTORNEY_SIGN",
				Changes: []shared.Change{
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/mobile",
						New: json.RawMessage(`"07777"`),
						Old: json.RawMessage(`"06666"`),
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/signedAt",
						New: json.RawMessage(`"` + now.Format(time.RFC3339Nano) + `"`),
						Old: json.RawMessage(`"` + yesterday.Format(time.RFC3339Nano) + `"`),
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/contactLanguagePreference",
						New: json.RawMessage(`"cy"`),
						Old: jsonNull,
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/channel",
						New: json.RawMessage(`"online"`),
						Old: json.RawMessage(`"paper"`),
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/email",
						New: json.RawMessage(`"b@example.com"`),
						Old: json.RawMessage(`"a@example.com"`),
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{}, {Person: shared.Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"}, Channel: shared.ChannelPaper, Email: "a@example.com", Mobile: "06666", SignedAt: &yesterday},
			}}},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			_, errors := validateUpdate(tc.update, tc.lpa)
			assert.ElementsMatch(t, tc.errors, errors)
		})
	}
}
