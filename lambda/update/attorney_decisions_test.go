package main

import (
	"encoding/json"
	"testing"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestAttorneyDecisionsApply(t *testing.T) {
	attorneyIndex0 := 0
	attorneyIndex1 := 1
	lpa := &shared.Lpa{
		Status: shared.LpaStatusInProgress,
		LpaInit: shared.LpaInit{
			Attorneys: []shared.Attorney{
				{Person: shared.Person{UID: "a"}, AppointmentType: shared.AppointmentTypeOriginal},
				{Person: shared.Person{UID: "b"}, AppointmentType: shared.AppointmentTypeReplacement},
			},
			HowAttorneysMakeDecisions:            shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
			HowReplacementAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
		},
	}

	c := AttorneyDecision{
		ChangeAttorneyDecisions: []ChangeAttorneyDecisions{
			{
				Index:     &attorneyIndex0,
				Decisions: true,
			},
			{
				Index:     &attorneyIndex1,
				Decisions: false,
			},
		},
	}

	errors := c.Apply(lpa)
	assert.Empty(t, errors)
	assert.Equal(t, c.ChangeAttorneyDecisions[0].Decisions, lpa.LpaInit.Attorneys[0].CannotMakeJointDecisions)
	assert.Equal(t, c.ChangeAttorneyDecisions[1].Decisions, lpa.LpaInit.Attorneys[1].CannotMakeJointDecisions)
}

func TestAttorneyDecisionsApplyErrors(t *testing.T) {

	testcases := map[string]struct {
		howMakeDecisions shared.HowMakeDecisions
		appointmentType  shared.AppointmentType
	}{
		"Sole - Original attorney": {
			howMakeDecisions: shared.HowMakeDecisionsUnset,
			appointmentType:  shared.AppointmentTypeOriginal,
		},
		"Sole - Replacement attorney": {
			howMakeDecisions: shared.HowMakeDecisionsUnset,
			appointmentType:  shared.AppointmentTypeReplacement,
		},
		"Jointly - Original attorney": {
			howMakeDecisions: shared.HowMakeDecisionsJointly,
			appointmentType:  shared.AppointmentTypeOriginal,
		},
		"Jointly - Replacement attorney": {
			howMakeDecisions: shared.HowMakeDecisionsJointly,
			appointmentType:  shared.AppointmentTypeReplacement,
		},
		"Jointly and severally - Original attorney": {
			howMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
			appointmentType:  shared.AppointmentTypeOriginal,
		},
		"Jointly and severally - Replacement attorney": {
			howMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
			appointmentType:  shared.AppointmentTypeReplacement,
		},
	}

	for scenario, tc := range testcases {
		attorneyIndex0 := 0
		lpa := &shared.Lpa{
			Status: shared.LpaStatusInProgress,
			LpaInit: shared.LpaInit{
				Attorneys: []shared.Attorney{
					{Person: shared.Person{UID: "a"}, AppointmentType: tc.appointmentType},
				},
				HowAttorneysMakeDecisions:            tc.howMakeDecisions,
				HowReplacementAttorneysMakeDecisions: tc.howMakeDecisions,
			},
		}

		c := AttorneyDecision{
			ChangeAttorneyDecisions: []ChangeAttorneyDecisions{
				{
					Index:     &attorneyIndex0,
					Decisions: true,
				},
			},
		}

		t.Run(scenario, func(t *testing.T) {
			errors := c.Apply(lpa)
			assert.Equal(t, errors,
				[]shared.FieldError{
					{
						Source: "/attorneys/0/cannotMakeJointDecisions",
						Detail: "The appointment type must be jointly for some and severally for others",
					},
				})
		})
	}
}

func TestValidateDecisions(t *testing.T) {

	update := shared.Update{
		Type: "ATTORNEY_DECISIONS",
		Changes: []shared.Change{
			{
				Key: "/attorneys/0/cannotMakeJointDecisions",
				New: json.RawMessage("true"),
				Old: json.RawMessage("false"),
			},
		},
	}

	lpa := &shared.Lpa{
		Status: shared.LpaStatusInProgress,
		LpaInit: shared.LpaInit{
			Attorneys: []shared.Attorney{
				{Person: shared.Person{UID: "a"}, AppointmentType: shared.AppointmentTypeOriginal},
				{Person: shared.Person{UID: "b"}, AppointmentType: shared.AppointmentTypeReplacement},
			},
			HowAttorneysMakeDecisions:            shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
			HowReplacementAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
		},
	}

	_, errors := validateUpdate(update, lpa)
	assert.ElementsMatch(t, errors, errors)
}
