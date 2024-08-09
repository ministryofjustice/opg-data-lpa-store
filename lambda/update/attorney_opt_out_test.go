package main

import (
	"encoding/json"
	"testing"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestAttorneyOptOutApply(t *testing.T) {
	testcases := map[string]struct {
		lpa         *shared.Lpa
		expectedLpa *shared.Lpa
	}{
		"single attorney": {
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsUnset,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusActive},
					},
				},
			},
			expectedLpa: &shared.Lpa{
				Status: shared.LpaStatusCannotRegister,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsUnset,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusRemoved},
					},
				},
			},
		},
		"two attorneys": {
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "a"}, Status: shared.AttorneyStatusActive},
						{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusActive},
					},
				},
			},
			expectedLpa: &shared.Lpa{
				Status: shared.LpaStatusCannotRegister,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "a"}, Status: shared.AttorneyStatusActive},
						{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusRemoved},
					},
				},
			},
		},
		"multiple attorneys jointly and severally": {
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "a"}, Status: shared.AttorneyStatusActive},
						{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusActive},
						{Person: shared.Person{UID: "c"}, Status: shared.AttorneyStatusActive},
					},
				},
			},
			expectedLpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "a"}, Status: shared.AttorneyStatusActive},
						{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusRemoved},
						{Person: shared.Person{UID: "c"}, Status: shared.AttorneyStatusActive},
					},
				},
			},
		},
		"multiple attorneys jointly": {
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointly,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "a"}, Status: shared.AttorneyStatusActive},
						{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusActive},
						{Person: shared.Person{UID: "c"}, Status: shared.AttorneyStatusActive},
					},
				},
			},
			expectedLpa: &shared.Lpa{
				Status: shared.LpaStatusCannotRegister,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointly,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "a"}, Status: shared.AttorneyStatusActive},
						{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusRemoved},
						{Person: shared.Person{UID: "c"}, Status: shared.AttorneyStatusActive},
					},
				},
			},
		},
		"multiple attorneys jointly for some": {
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "a"}, Status: shared.AttorneyStatusActive},
						{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusActive},
						{Person: shared.Person{UID: "c"}, Status: shared.AttorneyStatusActive},
					},
				},
			},
			expectedLpa: &shared.Lpa{
				Status: shared.LpaStatusCannotRegister,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "a"}, Status: shared.AttorneyStatusActive},
						{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusRemoved},
						{Person: shared.Person{UID: "c"}, Status: shared.AttorneyStatusActive},
					},
				},
			},
		},
		"multiple attorneys with trust corporations": {
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "a"}, Status: shared.AttorneyStatusActive},
						{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusActive},
					},
					TrustCorporations: []shared.TrustCorporation{
						{Status: shared.AttorneyStatusActive},
					},
				},
			},
			expectedLpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "a"}, Status: shared.AttorneyStatusActive},
						{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusRemoved},
					},
					TrustCorporations: []shared.TrustCorporation{
						{Status: shared.AttorneyStatusActive},
					},
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			c := AttorneyOptOut{AttorneyUID: "b"}

			errors := c.Apply(tc.lpa)

			assert.Empty(t, errors)
			assert.Equal(t, tc.expectedLpa, tc.lpa)
		})
	}
}

func TestValidateUpdateAttorneyOptOut(t *testing.T) {
	testcases := map[string]struct {
		update   shared.Update
		errors   []shared.FieldError
		expected AttorneyOptOut
	}{
		"valid": {
			update: shared.Update{
				Author:  "urn:opg:poas:makeregister:users:dc487ebb-b39d-45ed-bb6a-7f950fd355c9",
				Type:    "ATTORNEY_OPT_OUT",
				Changes: []shared.Change{},
				Subject: "dc487ebb-b39d-45ed-bb6a-7f950fd355c9",
			},
			expected: AttorneyOptOut{AttorneyUID: "dc487ebb-b39d-45ed-bb6a-7f950fd355c9"},
		},
		"with changes": {
			update: shared.Update{
				Author: "urn:opg:poas:makeregister:users:dc487ebb-b39d-45ed-bb6a-7f950fd355c9",
				Type:   "ATTORNEY_OPT_OUT",
				Changes: []shared.Change{
					{
						Key: "/something/someValue",
						New: json.RawMessage(`"not expected"`),
						Old: jsonNull,
					},
				},
				Subject: "dc487ebb-b39d-45ed-bb6a-7f950fd355c9",
			},
			expected: AttorneyOptOut{},
			errors: []shared.FieldError{
				{Source: "/changes", Detail: "expected empty"},
			},
		},
		"invalid subject": {
			update: shared.Update{
				Author:  "urn:opg:poas:makeregister:users:dc487ebb-b39d-45ed-bb6a-7f950fd355c9",
				Type:    "ATTORNEY_OPT_OUT",
				Changes: []shared.Change{},
				Subject: "123",
			},
			expected: AttorneyOptOut{},
			errors: []shared.FieldError{
				{Source: "/subject", Detail: "invalid format"},
			},
		},
		"missing subject": {
			update: shared.Update{
				Author:  "urn:opg:poas:makeregister:users:dc487ebb-b39d-45ed-bb6a-7f950fd355c9",
				Type:    "ATTORNEY_OPT_OUT",
				Changes: []shared.Change{},
				Subject: "",
			},
			expected: AttorneyOptOut{},
			errors: []shared.FieldError{
				{Source: "/subject", Detail: "field is required"},
			},
		},
		"subject and author do not match": {
			update: shared.Update{
				Author:  "urn:opg:poas:makeregister:users:dc487ebb-b39d-45ed-bb6a-7f950fd355c9",
				Type:    "ATTORNEY_OPT_OUT",
				Changes: []shared.Change{},
				Subject: "dc487ebb-b39d-45ed-bb6a-7f950fd355c8",
			},
			expected: AttorneyOptOut{},
			errors: []shared.FieldError{
				{Source: "/update", Detail: "cannot change other actors"},
			},
		},
		"subject and author do not match when service not makeregister": {
			update: shared.Update{
				Author:  "urn:opg:poas:sirius:users:dc487ebb-b39d-45ed-bb6a-7f950fd355c9",
				Type:    "ATTORNEY_OPT_OUT",
				Changes: []shared.Change{},
				Subject: "dc487ebb-b39d-45ed-bb6a-7f950fd355c8",
			},
			expected: AttorneyOptOut{AttorneyUID: "dc487ebb-b39d-45ed-bb6a-7f950fd355c8"},
			errors:   []shared.FieldError{},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			data, errors := validateUpdate(tc.update, &shared.Lpa{})
			assert.Equal(t, tc.expected, data)
			assert.ElementsMatch(t, tc.errors, errors)
		})
	}
}
