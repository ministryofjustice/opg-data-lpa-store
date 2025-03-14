package main

import "github.com/ministryofjustice/opg-data-lpa-store/internal/shared"

func SetDefaults(input shared.LpaInit) shared.LpaInit {
	activeAttorneyCount, replacementAttorneyCount := shared.CountAttorneys(input.Attorneys, input.TrustCorporations)

	if activeAttorneyCount > 1 && input.HowAttorneysMakeDecisions.Unset() {
		input.HowAttorneysMakeDecisions = shared.HowMakeDecisionsJointly
		input.HowAttorneysMakeDecisionsIsDefault = true
	}

	shouldHaveReplacementDecisions := replacementAttorneyCount > 1 && (input.HowReplacementAttorneysStepIn == shared.HowStepInAllCanNoLongerAct ||
		input.HowAttorneysMakeDecisions != shared.HowMakeDecisionsJointlyAndSeverally)

	if shouldHaveReplacementDecisions && input.HowReplacementAttorneysMakeDecisions.Unset() {
		input.HowReplacementAttorneysMakeDecisions = shared.HowMakeDecisionsJointly
		input.HowReplacementAttorneysMakeDecisionsIsDefault = true
	}

	if input.LpaType == shared.LpaTypePropertyAndAffairs && input.WhenTheLpaCanBeUsed.Unset() {
		input.WhenTheLpaCanBeUsed = shared.CanUseWhenHasCapacity
		input.WhenTheLpaCanBeUsedIsDefault = true
	}

	if input.LpaType == shared.LpaTypePersonalWelfare && input.LifeSustainingTreatmentOption.Unset() {
		input.LifeSustainingTreatmentOption = shared.LifeSustainingTreatmentOptionB
		input.LifeSustainingTreatmentOptionIsDefault = true
	}

	return input
}
