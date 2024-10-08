package main

import (
	"fmt"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
)

func Validate(lpa shared.LpaInit) []shared.FieldError {
	activeAttorneyCount, replacementAttorneyCount := countAttorneys(lpa.Attorneys, lpa.TrustCorporations)

	return validate.All(
		validate.IsValid("/lpaType", lpa.LpaType),
		validate.IsValid("/channel", lpa.Channel),
		validate.UUID("/donor/uid", lpa.Donor.UID),
		validate.Required("/donor/firstNames", lpa.Donor.FirstNames),
		validate.Required("/donor/lastName", lpa.Donor.LastName),
		validate.Date("/donor/dateOfBirth", lpa.Donor.DateOfBirth),
		validate.Address("/donor/address", lpa.Donor.Address),
		validate.IsValid("/donor/contactLanguagePreference", lpa.Donor.ContactLanguagePreference),
		validate.IfFunc(lpa.Donor.IdentityCheck != nil, func() []shared.FieldError {
			return validate.All(
				validate.Time("/donor/identityCheck/checkedAt", lpa.Donor.IdentityCheck.CheckedAt),
				validate.IsValid("/donor/identityCheck/type", lpa.Donor.IdentityCheck.Type))
		}),
		validate.UUID("/certificateProvider/uid", lpa.CertificateProvider.UID),
		validate.Required("/certificateProvider/firstNames", lpa.CertificateProvider.FirstNames),
		validate.Required("/certificateProvider/lastName", lpa.CertificateProvider.LastName),
		validate.Address("/certificateProvider/address", lpa.CertificateProvider.Address),
		validate.IsValid("/certificateProvider/channel", lpa.CertificateProvider.Channel),
		validate.IfElse(lpa.CertificateProvider.Channel == shared.ChannelOnline,
			validate.Required("/certificateProvider/email", lpa.CertificateProvider.Email),
			validate.Empty("/certificateProvider/email", lpa.CertificateProvider.Email)),
		validate.Required("/certificateProvider/phone", lpa.CertificateProvider.Phone),
		validateAttorneys("/attorneys", lpa.Attorneys),
		validateTrustCorporations("/trustCorporations", lpa.TrustCorporations),
		validate.IfFunc(lpa.AuthorisedSignatory != nil, func() []shared.FieldError {
			return validate.All(
				validate.Required("/authorisedSignatory/uid", lpa.AuthorisedSignatory.UID),
				validate.Required("/authorisedSignatory/firstNames", lpa.AuthorisedSignatory.FirstNames),
				validate.Required("/authorisedSignatory/lastName", lpa.AuthorisedSignatory.LastName))
		}),
		validate.IfFunc(lpa.IndependentWitness != nil, func() []shared.FieldError {
			return validate.All(
				validate.Required("/independentWitness/uid", lpa.IndependentWitness.UID),
				validate.Required("/independentWitness/firstNames", lpa.IndependentWitness.FirstNames),
				validate.Required("/independentWitness/lastName", lpa.IndependentWitness.LastName),
				validate.Required("/independentWitness/phone", lpa.IndependentWitness.Phone),
				validate.Address("/independentWitness/address", lpa.IndependentWitness.Address))
		}),
		validate.IfElse(activeAttorneyCount > 1,
			validate.IsValid("/howAttorneysMakeDecisions", lpa.HowAttorneysMakeDecisions),
			validate.Unset("/howAttorneysMakeDecisions", lpa.HowAttorneysMakeDecisions)),
		validate.IfElse(lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
			validate.Required("/howAttorneysMakeDecisionsDetails", lpa.HowAttorneysMakeDecisionsDetails),
			validate.Empty("/howAttorneysMakeDecisionsDetails", lpa.HowAttorneysMakeDecisionsDetails)),
		validate.If(replacementAttorneyCount > 0 && lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyAndSeverally,
			validate.IsValid("/howReplacementAttorneysStepIn", lpa.HowReplacementAttorneysStepIn)),
		validate.IfElse(lpa.HowReplacementAttorneysStepIn == shared.HowStepInAnotherWay,
			validate.Required("/howReplacementAttorneysStepInDetails", lpa.HowReplacementAttorneysStepInDetails),
			validate.Empty("/howReplacementAttorneysStepInDetails", lpa.HowReplacementAttorneysStepInDetails)),
		validate.IfElse(replacementAttorneyCount > 1 && (lpa.HowReplacementAttorneysStepIn == shared.HowStepInAllCanNoLongerAct || lpa.HowAttorneysMakeDecisions != shared.HowMakeDecisionsJointlyAndSeverally),
			validate.IsValid("/howReplacementAttorneysMakeDecisions", lpa.HowReplacementAttorneysMakeDecisions),
			validate.Unset("/howReplacementAttorneysMakeDecisions", lpa.HowReplacementAttorneysMakeDecisions)),
		validate.IfElse(lpa.HowReplacementAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
			validate.Required("/howReplacementAttorneysMakeDecisionsDetails", lpa.HowReplacementAttorneysMakeDecisionsDetails),
			validate.Empty("/howReplacementAttorneysMakeDecisionsDetails", lpa.HowReplacementAttorneysMakeDecisionsDetails)),
		validate.If(lpa.LpaType == shared.LpaTypePersonalWelfare, validate.All(
			validate.IsValid("/lifeSustainingTreatmentOption", lpa.LifeSustainingTreatmentOption),
			validate.Unset("/whenTheLpaCanBeUsed", lpa.WhenTheLpaCanBeUsed))),
		validate.If(lpa.LpaType == shared.LpaTypePropertyAndAffairs, validate.All(
			validate.IsValid("/whenTheLpaCanBeUsed", lpa.WhenTheLpaCanBeUsed),
			validate.Unset("/lifeSustainingTreatmentOption", lpa.LifeSustainingTreatmentOption))),
		validate.Time("/signedAt", lpa.SignedAt),
		// validate.Time("/witnessedByCertificateProviderAt", lpa.WitnessedByCertificateProviderAt),
		// validate.OptionalTime("/witnessedByIndependentWitnessAt", lpa.WitnessedByIndependentWitnessAt),
		validate.OptionalTime("/certificateProviderNotRelatedConfirmedAt", lpa.CertificateProviderNotRelatedConfirmedAt),
	)
}

func countAttorneys(as []shared.Attorney, ts []shared.TrustCorporation) (actives, replacements int) {
	for _, a := range as {
		switch a.Status {
		case shared.AttorneyStatusActive:
			actives++
		case shared.AttorneyStatusReplacement:
			replacements++
		}
	}

	for _, t := range ts {
		switch t.Status {
		case shared.AttorneyStatusActive:
			actives++
		case shared.AttorneyStatusReplacement:
			replacements++
		}
	}

	return actives, replacements
}

func validateAttorneys(prefix string, attorneys []shared.Attorney) []shared.FieldError {
	var errors []shared.FieldError

	if len(attorneys) == 0 {
		return []shared.FieldError{{Source: prefix, Detail: "at least one attorney is required"}}
	}

	for i, attorney := range attorneys {
		if e := validateAttorney(fmt.Sprintf("%s/%d", prefix, i), attorney); e != nil {
			errors = append(errors, e...)
		}
	}

	return errors
}

func validateAttorney(prefix string, attorney shared.Attorney) []shared.FieldError {
	return validate.All(
		validate.UUID(fmt.Sprintf("%s/uid", prefix), attorney.UID),
		validate.Required(fmt.Sprintf("%s/firstNames", prefix), attorney.FirstNames),
		validate.Required(fmt.Sprintf("%s/lastName", prefix), attorney.LastName),
		validate.Date(fmt.Sprintf("%s/dateOfBirth", prefix), attorney.DateOfBirth),
		validate.Address(fmt.Sprintf("%s/address", prefix), attorney.Address),
		validate.IsValid(fmt.Sprintf("%s/status", prefix), attorney.Status),
		validate.IsValid(fmt.Sprintf("%s/channel", prefix), attorney.Channel),
		validate.IfElse(attorney.Channel == shared.ChannelOnline,
			validate.Required(fmt.Sprintf("%s/email", prefix), attorney.Email),
			validate.Empty(fmt.Sprintf("%s/email", prefix), attorney.Email)),
	)
}

func validateTrustCorporations(prefix string, trustCorporations []shared.TrustCorporation) []shared.FieldError {
	var errors []shared.FieldError

	for i, trustCorporation := range trustCorporations {
		if e := validateTrustCorporation(fmt.Sprintf("%s/%d", prefix, i), trustCorporation); e != nil {
			errors = append(errors, e...)
		}
	}

	return errors
}

func validateTrustCorporation(prefix string, trustCorporation shared.TrustCorporation) []shared.FieldError {
	return validate.All(
		validate.UUID(fmt.Sprintf("%s/uid", prefix), trustCorporation.UID),
		validate.Required(fmt.Sprintf("%s/name", prefix), trustCorporation.Name),
		validate.Required(fmt.Sprintf("%s/companyNumber", prefix), trustCorporation.CompanyNumber),
		validate.Address(fmt.Sprintf("%s/address", prefix), trustCorporation.Address),
		validate.IsValid(fmt.Sprintf("%s/status", prefix), trustCorporation.Status),
		validate.IsValid(fmt.Sprintf("%s/channel", prefix), trustCorporation.Channel),
		validate.IfElse(trustCorporation.Channel == shared.ChannelOnline,
			validate.Required(fmt.Sprintf("%s/email", prefix), trustCorporation.Email),
			validate.Empty(fmt.Sprintf("%s/email", prefix), trustCorporation.Email)),
	)
}
