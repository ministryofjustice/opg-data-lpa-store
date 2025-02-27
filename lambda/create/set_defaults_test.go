package main

import (
	"testing"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestSetDefaults(t *testing.T) {
	testcases := map[string]struct {
		in, out shared.LpaInit
	}{
		"empty": {},
		"how attorneys make decisions": {
			in: shared.LpaInit{
				Attorneys: []shared.Attorney{{Status: shared.AttorneyStatusActive}, {Status: shared.AttorneyStatusActive}},
			},
			out: shared.LpaInit{
				Attorneys:                          []shared.Attorney{{Status: shared.AttorneyStatusActive}, {Status: shared.AttorneyStatusActive}},
				HowAttorneysMakeDecisions:          shared.HowMakeDecisionsJointly,
				HowAttorneysMakeDecisionsIsDefault: true,
			},
		},
		"how replacement attorneys make decisions should not be set": {
			in: shared.LpaInit{
				Attorneys: []shared.Attorney{
					{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
					{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
				},
				HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
			},
			out: shared.LpaInit{
				Attorneys: []shared.Attorney{
					{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
					{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
				},
				HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
			},
		},
		"how replacement attorneys make decisions from attorneys jointly": {
			in: shared.LpaInit{
				Attorneys: []shared.Attorney{
					{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
					{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
				},
				HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointly,
			},
			out: shared.LpaInit{
				Attorneys: []shared.Attorney{
					{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
					{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
				},
				HowAttorneysMakeDecisions:                     shared.HowMakeDecisionsJointly,
				HowReplacementAttorneysMakeDecisions:          shared.HowMakeDecisionsJointly,
				HowReplacementAttorneysMakeDecisionsIsDefault: true,
			},
		},
		"how replacement attorneys make decisions from step in": {
			in: shared.LpaInit{
				Attorneys: []shared.Attorney{
					{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
					{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
				},
				HowReplacementAttorneysStepIn: shared.HowStepInAllCanNoLongerAct,
			},
			out: shared.LpaInit{
				Attorneys: []shared.Attorney{
					{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
					{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
				},
				HowReplacementAttorneysStepIn:                 shared.HowStepInAllCanNoLongerAct,
				HowReplacementAttorneysMakeDecisions:          shared.HowMakeDecisionsJointly,
				HowReplacementAttorneysMakeDecisionsIsDefault: true,
			},
		},
		"when the lpa can be used": {
			in: shared.LpaInit{
				LpaType: shared.LpaTypePropertyAndAffairs,
			},
			out: shared.LpaInit{
				LpaType:                      shared.LpaTypePropertyAndAffairs,
				WhenTheLpaCanBeUsed:          shared.CanUseWhenHasCapacity,
				WhenTheLpaCanBeUsedIsDefault: true,
			},
		},
		"life sustaining treatment": {
			in: shared.LpaInit{
				LpaType: shared.LpaTypePersonalWelfare,
			},
			out: shared.LpaInit{
				LpaType:                                shared.LpaTypePersonalWelfare,
				LifeSustainingTreatmentOption:          shared.LifeSustainingTreatmentOptionB,
				LifeSustainingTreatmentOptionIsDefault: true,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.out, SetDefaults(tc.in))
		})
	}
}
