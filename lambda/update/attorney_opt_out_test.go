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
						{Person: shared.Person{UID: "b"}},
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
						{Person: shared.Person{UID: "a"}},
						{Person: shared.Person{UID: "b"}},
					},
				},
			},
			expectedLpa: &shared.Lpa{
				Status: shared.LpaStatusCannotRegister,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "a"}},
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
						{Person: shared.Person{UID: "a"}},
						{Person: shared.Person{UID: "b"}},
						{Person: shared.Person{UID: "c"}},
					},
				},
			},
			expectedLpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "a"}},
						{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusRemoved},
						{Person: shared.Person{UID: "c"}},
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
						{Person: shared.Person{UID: "a"}},
						{Person: shared.Person{UID: "b"}},
						{Person: shared.Person{UID: "c"}},
					},
				},
			},
			expectedLpa: &shared.Lpa{
				Status: shared.LpaStatusCannotRegister,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointly,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "a"}},
						{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusRemoved},
						{Person: shared.Person{UID: "c"}},
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
						{Person: shared.Person{UID: "a"}},
						{Person: shared.Person{UID: "b"}},
						{Person: shared.Person{UID: "c"}},
					},
				},
			},
			expectedLpa: &shared.Lpa{
				Status: shared.LpaStatusCannotRegister,
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "a"}},
						{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusRemoved},
						{Person: shared.Person{UID: "c"}},
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
				Author:  "urn:opg:poas:makeregister:users:123",
				Type:    "ATTORNEY_OPT_OUT",
				Changes: []shared.Change{},
			},
			expected: AttorneyOptOut{AttorneyUID: "123"},
		},
		"with changes": {
			update: shared.Update{
				Author: "urn:opg:poas:makeregister:users:123",
				Type:   "ATTORNEY_OPT_OUT",
				Changes: []shared.Change{
					{
						Key: "/something/someValue",
						New: json.RawMessage(`"not expected"`),
						Old: jsonNull,
					},
				},
			},
			expected: AttorneyOptOut{},
			errors: []shared.FieldError{
				{Source: "/changes", Detail: "expected empty"},
			},
		},
		"invalid author UID": {
			update: shared.Update{
				Author:  "urn:opg:poas:makeregister:users:",
				Type:    "ATTORNEY_OPT_OUT",
				Changes: []shared.Change{},
			},
			expected: AttorneyOptOut{},
			errors: []shared.FieldError{
				{Source: "/update", Detail: "author UID missing from URN"},
			},
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
