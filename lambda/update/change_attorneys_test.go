package main

import (
	"encoding/json"
	"testing"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestChangeAttorneysApply(t *testing.T) {
	attorneyIndex := 1
	lpa := &shared.Lpa{
		LpaInit: shared.LpaInit{
			Attorneys: []shared.Attorney{
				{}, {},
			},
		},
	}

	changeAttorney := ChangeAttorney{
		ChangeAttorneyStatus: []ChangeAttorneyStatus{
			{
				Index:  &attorneyIndex,
				Status: shared.AttorneyStatusActive,
			},
		},
	}

	errors := changeAttorney.Apply(lpa)
	assert.Empty(t, errors)
	assert.Equal(t, changeAttorney.ChangeAttorneyStatus[0].Status, lpa.Attorneys[attorneyIndex].Status)
	assert.Empty(t, lpa.Notes)
}

func TestChangeAttorneysApplySetAttorneyInactive(t *testing.T) {
	attorneyIndex := 1
	lpa := &shared.Lpa{
		LpaInit: shared.LpaInit{
			Attorneys: []shared.Attorney{
				{Person: shared.Person{
					FirstNames: "Arun",
					LastName:   "Brar",
				}},
				{Person: shared.Person{
					FirstNames: "Charles",
					LastName:   "Dent",
				}},
			},
		},
	}

	changeAttorney := ChangeAttorney{
		ChangeAttorneyStatus: []ChangeAttorneyStatus{
			{
				Index:  &attorneyIndex,
				Status: shared.AttorneyStatusRemoved,
			},
		},
	}

	errors := changeAttorney.Apply(lpa)

	noteValues := lpa.Notes[0]["values"].(map[string]string)

	assert.Empty(t, errors)
	assert.Equal(t, changeAttorney.ChangeAttorneyStatus[0].Status, lpa.Attorneys[attorneyIndex].Status)
	assert.Len(t, lpa.Notes, 1)
	assert.Equal(t, "ATTORNEY_REMOVED_V1", lpa.Notes[0]["type"])
	assert.Equal(t, "Charles Dent", noteValues["fullName"])
}

func TestChangeAttorneysIncorrectStatus(t *testing.T) {
	attorneyIndex0 := 0
	lpa := &shared.Lpa{
		LpaInit: shared.LpaInit{
			Attorneys: []shared.Attorney{
				{Status: shared.AttorneyStatusActive}, {Status: shared.AttorneyStatusInactive},
			},
		},
	}

	changeAttorney := ChangeAttorney{
		ChangeAttorneyStatus: []ChangeAttorneyStatus{
			{
				Index:  &attorneyIndex0,
				Status: shared.AttorneyStatusInactive,
			},
		},
	}

	errors := changeAttorney.Apply(lpa)
	assert.Equal(t, errors,
		[]shared.FieldError{
			{
				Source: "/attorneys/0/status",
				Detail: "An active attorney cannot be made inactive",
			},
		})
}

func TestValidateUpdateChangeAttorneys(t *testing.T) {
	testcases := map[string]struct {
		update shared.Update
		lpa    *shared.Lpa
		errors []shared.FieldError
	}{
		"valid - with previous values": {
			update: shared.Update{
				Type: "CHANGE_ATTORNEYS",
				Changes: []shared.Change{
					{
						Key: "/attorneys/1/status",
						New: json.RawMessage(`"removed"`),
						Old: json.RawMessage(`"active"`),
					},
					{
						Key: "/attorneys/2/status",
						New: json.RawMessage(`"active"`),
						Old: json.RawMessage(`"inactive"`),
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{Status: shared.AttorneyStatusActive}, {Status: shared.AttorneyStatusActive}, {Status: shared.AttorneyStatusInactive},
			}}},
		},
		"invalid status": {
			update: shared.Update{
				Type: "CHANGE_ATTORNEYS",
				Changes: []shared.Change{
					{
						Key: "/attorneys/0/status",
						New: json.RawMessage(`"in-progress"`),
						Old: json.RawMessage(`"active"`),
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{Status: shared.AttorneyStatusActive}, {Status: shared.AttorneyStatusInactive},
			}}},
			errors: []shared.FieldError{
				{Source: "/changes/0/new", Detail: "invalid value"},
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

func TestValidateUpdateChangeAttorneysWithUIDReferences(t *testing.T) {
	testcases := map[string]struct {
		update shared.Update
		lpa    *shared.Lpa
		errors []shared.FieldError
	}{
		"valid - with previous values": {
			update: shared.Update{
				Type: "CHANGE_ATTORNEYS",
				Changes: []shared.Change{
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/status",
						New: json.RawMessage(`"removed"`),
						Old: json.RawMessage(`"active"`),
					},
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3e/status",
						New: json.RawMessage(`"active"`),
						Old: json.RawMessage(`"inactive"`),
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{Person: shared.Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3c"}, Status: shared.AttorneyStatusActive},
				{Person: shared.Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"}, Status: shared.AttorneyStatusActive},
				{Person: shared.Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3e"}, Status: shared.AttorneyStatusInactive},
			}}},
		},
		"invalid status": {
			update: shared.Update{
				Type: "CHANGE_ATTORNEYS",
				Changes: []shared.Change{
					{
						Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/status",
						New: json.RawMessage(`"in-progress"`),
						Old: json.RawMessage(`"active"`),
					},
				},
			},
			lpa: &shared.Lpa{LpaInit: shared.LpaInit{Attorneys: []shared.Attorney{
				{Person: shared.Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"}, Status: shared.AttorneyStatusActive},
				{Status: shared.AttorneyStatusInactive},
			}}},
			errors: []shared.FieldError{
				{Source: "/changes/0/new", Detail: "invalid value"},
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

func TestChangeAttorneysEnableReplacementAttorney(t *testing.T) {
	attorneyIndex := 1
	lpa := &shared.Lpa{
		LpaInit: shared.LpaInit{
			Attorneys: []shared.Attorney{
				{
					Person: shared.Person{
						FirstNames: "Arun",
						LastName:   "Brar",
					},
				},
				{
					Person: shared.Person{
						FirstNames: "Charles",
						LastName:   "Dent",
					},
					AppointmentType: shared.AppointmentTypeReplacement,
				},
			},
		},
	}

	changeAttorney := ChangeAttorney{
		ChangeAttorneyStatus: []ChangeAttorneyStatus{
			{
				Index:  &attorneyIndex,
				Status: shared.AttorneyStatusActive,
			},
		},
	}

	errors := changeAttorney.Apply(lpa)

	noteValues := lpa.Notes[0]["values"].(map[string]string)

	assert.Empty(t, errors)
	assert.Equal(t, changeAttorney.ChangeAttorneyStatus[0].Status, lpa.Attorneys[attorneyIndex].Status)
	assert.Len(t, lpa.Notes, 1)
	assert.Equal(t, "REPLACEMENT_ATTORNEY_ENABLED_V1", lpa.Notes[0]["type"])
	assert.Equal(t, "Charles Dent", noteValues["fullName"])
}
