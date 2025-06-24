package main

import (
	"encoding/json"
	"testing"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestAttorneyDecisionsApply(t *testing.T) {

	testcases := map[string]struct {
		appointmentType shared.HowMakeDecisions
	}{
		"Jointly for some severally for others": {
			appointmentType: shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
		},
	}

	for scenario, tc := range testcases {
		attorneyIndex0 := 0
		attorneyIndex1 := 1
		lpa := &shared.Lpa{
			Status: shared.LpaStatusInProgress,
			LpaInit: shared.LpaInit{
				Attorneys: []shared.Attorney{
					{Person: shared.Person{UID: "a"}, Status: shared.AttorneyStatusActive, AppointmentType: shared.AppointmentTypeOriginal},
					{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusActive, AppointmentType: shared.AppointmentTypeReplacement},
					{Person: shared.Person{UID: "c"}, Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
				},
				HowAttorneysMakeDecisions:            tc.appointmentType,
				HowReplacementAttorneysMakeDecisions: tc.appointmentType,
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

		t.Run(scenario, func(t *testing.T) {
			errors := c.Apply(lpa)
			assert.Empty(t, errors)
			assert.Equal(t, c.ChangeAttorneyDecisions[0].Decisions, lpa.LpaInit.Attorneys[0].CannotMakeJointDecisions)
			assert.Equal(t, c.ChangeAttorneyDecisions[1].Decisions, lpa.LpaInit.Attorneys[1].CannotMakeJointDecisions)
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
				{Person: shared.Person{UID: "a"}, Status: shared.AttorneyStatusActive, AppointmentType: shared.AppointmentTypeOriginal},
				{Person: shared.Person{UID: "b"}, Status: shared.AttorneyStatusActive, AppointmentType: shared.AppointmentTypeReplacement},
				{Person: shared.Person{UID: "c"}, Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
			},
			HowAttorneysMakeDecisions:            shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
			HowReplacementAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
		},
	}

	_, errors := validateUpdate(update, lpa)
	assert.ElementsMatch(t, errors, errors)
}
