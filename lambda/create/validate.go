package main

import (
	"fmt"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
)

func validateAddress(prefix string, address shared.Address) []shared.FieldError {
	return validate.All(
		validate.WithSource(prefix+"/line1", address.Line1, validate.NotEmpty()),
		validate.WithSource(prefix+"/country", address.Country, validate.NotEmpty(), validate.Country()),
	)
}

func Validate(lpa shared.LpaInit) []shared.FieldError {
	activeAttorneyCount, replacementAttorneyCount := shared.CountAttorneys(lpa.Attorneys, lpa.TrustCorporations)

	return validate.All(
		validate.WithSource("/lpaType", lpa.LpaType, validate.Valid()),
		validate.WithSource("/channel", lpa.Channel, validate.Valid()),
		validate.WithSource("/language", lpa.Language, validate.Valid()),
		validate.WithSource("/donor/uid", lpa.Donor.UID, validate.UUID()),
		validate.WithSource("/donor/firstNames", lpa.Donor.FirstNames, validate.NotEmpty()),
		validate.WithSource("/donor/lastName", lpa.Donor.LastName, validate.NotEmpty()),
		validate.WithSource("/donor/dateOfBirth", lpa.Donor.DateOfBirth, validate.Date()),
		validateAddress("/donor/address", lpa.Donor.Address),
		validate.WithSource("/donor/contactLanguagePreference", lpa.Donor.ContactLanguagePreference, validate.Valid()),
		validate.IfFunc(lpa.Donor.IdentityCheck != nil, func() []shared.FieldError {
			return validate.All(
				validate.WithSource("/donor/identityCheck/checkedAt", lpa.Donor.IdentityCheck.CheckedAt, validate.NotEmpty()),
				validate.WithSource("/donor/identityCheck/type", lpa.Donor.IdentityCheck.Type, validate.Valid()))
		}),
		validate.WithSource("/certificateProvider/uid", lpa.CertificateProvider.UID, validate.UUID()),
		validate.WithSource("/certificateProvider/firstNames", lpa.CertificateProvider.FirstNames, validate.NotEmpty()),
		validate.WithSource("/certificateProvider/lastName", lpa.CertificateProvider.LastName, validate.NotEmpty()),
		validateAddress("/certificateProvider/address", lpa.CertificateProvider.Address),
		validate.WithSource("/certificateProvider/channel", lpa.CertificateProvider.Channel, validate.Valid()),
		validate.IfElse(lpa.CertificateProvider.Channel == shared.ChannelOnline,
			validate.WithSource("/certificateProvider/email", lpa.CertificateProvider.Email, validate.NotEmpty()),
			validate.WithSource("/certificateProvider/email", lpa.CertificateProvider.Email, validate.Empty())),
		validate.WithSource("/certificateProvider/phone", lpa.CertificateProvider.Phone, validate.NotEmpty()),
		validateAttorneys("/attorneys", lpa.Attorneys),
		validateTrustCorporations("/trustCorporations", lpa.TrustCorporations),
		validate.IfFunc(lpa.AuthorisedSignatory != nil, func() []shared.FieldError {
			return validate.All(
				validate.WithSource("/authorisedSignatory/uid", lpa.AuthorisedSignatory.UID, validate.NotEmpty()),
				validate.WithSource("/authorisedSignatory/firstNames", lpa.AuthorisedSignatory.FirstNames, validate.NotEmpty()),
				validate.WithSource("/authorisedSignatory/lastName", lpa.AuthorisedSignatory.LastName, validate.NotEmpty()))
		}),
		validate.IfFunc(lpa.IndependentWitness != nil, func() []shared.FieldError {
			return validate.All(
				validate.WithSource("/independentWitness/uid", lpa.IndependentWitness.UID, validate.NotEmpty()),
				validate.WithSource("/independentWitness/firstNames", lpa.IndependentWitness.FirstNames, validate.NotEmpty()),
				validate.WithSource("/independentWitness/lastName", lpa.IndependentWitness.LastName, validate.NotEmpty()),
				validate.WithSource("/independentWitness/phone", lpa.IndependentWitness.Phone, validate.NotEmpty()),
				validateAddress("/independentWitness/address", lpa.IndependentWitness.Address),
				validate.WithSource("/witnessedByIndependentWitnessAt", lpa.WitnessedByIndependentWitnessAt, validate.NotEmpty()))
		}),
		validate.IfElse(activeAttorneyCount > 1,
			validate.WithSource("/howAttorneysMakeDecisions", lpa.HowAttorneysMakeDecisions, validate.Valid()),
			validate.WithSource("/howAttorneysMakeDecisions", lpa.HowAttorneysMakeDecisions, validate.Unset())),
		validate.IfElse(lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers && lpa.Channel == shared.ChannelOnline,
			validate.WithSource("/howAttorneysMakeDecisionsDetails", lpa.HowAttorneysMakeDecisionsDetails, validate.NotEmpty()),
			validate.IfElse(lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers && lpa.Channel == shared.ChannelPaper,
				validate.WithSource("/howAttorneysMakeDecisionsDetailsImages", lpa.HowAttorneysMakeDecisionsDetails, validate.NotEmpty()),
				validate.WithSource("/howAttorneysMakeDecisionsDetails", lpa.HowAttorneysMakeDecisionsDetails, validate.Empty()),
			),
		),
		validate.If(replacementAttorneyCount > 0 && lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyAndSeverally,
			validate.WithSource("/howReplacementAttorneysStepIn", lpa.HowReplacementAttorneysStepIn, validate.Valid())),
		validate.IfElse(lpa.HowReplacementAttorneysStepIn == shared.HowStepInAnotherWay,
			validate.WithSource("/howReplacementAttorneysStepInDetails", lpa.HowReplacementAttorneysStepInDetails, validate.NotEmpty()),
			validate.WithSource("/howReplacementAttorneysStepInDetails", lpa.HowReplacementAttorneysStepInDetails, validate.Empty())),
		validate.IfElse(replacementAttorneyCount > 1 && (activeAttorneyCount == 1 || lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointly || (lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyAndSeverally && lpa.HowReplacementAttorneysStepIn == shared.HowStepInAllCanNoLongerAct)),
			validate.WithSource("/howReplacementAttorneysMakeDecisions", lpa.HowReplacementAttorneysMakeDecisions, validate.Valid()),
			validate.WithSource("/howReplacementAttorneysMakeDecisions", lpa.HowReplacementAttorneysMakeDecisions, validate.Unset())),
		validate.IfElse(lpa.HowReplacementAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
			validate.WithSource("/howReplacementAttorneysMakeDecisionsDetails", lpa.HowReplacementAttorneysMakeDecisionsDetails, validate.NotEmpty()),
			validate.WithSource("/howReplacementAttorneysMakeDecisionsDetails", lpa.HowReplacementAttorneysMakeDecisionsDetails, validate.Empty())),
		validate.If(lpa.LpaType == shared.LpaTypePersonalWelfare, validate.All(
			validate.WithSource("/lifeSustainingTreatmentOption", lpa.LifeSustainingTreatmentOption, validate.Valid()),
			validate.WithSource("/whenTheLpaCanBeUsed", lpa.WhenTheLpaCanBeUsed, validate.Unset()))),
		validate.If(lpa.LpaType == shared.LpaTypePropertyAndAffairs, validate.All(
			validate.WithSource("/whenTheLpaCanBeUsed", lpa.WhenTheLpaCanBeUsed, validate.Valid()),
			validate.WithSource("/lifeSustainingTreatmentOption", lpa.LifeSustainingTreatmentOption, validate.Unset()))),
		validate.WithSource("/signedAt", lpa.SignedAt, validate.NotEmpty()),
		validate.WithSource("/witnessedByCertificateProviderAt", lpa.WitnessedByCertificateProviderAt, validate.NotEmpty()),
		validate.WithSource("/certificateProviderNotRelatedConfirmedAt", lpa.CertificateProviderNotRelatedConfirmedAt, validate.OptionalTime()),
	)
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
		validate.WithSource(fmt.Sprintf("%s/uid", prefix), attorney.UID, validate.UUID()),
		validate.WithSource(fmt.Sprintf("%s/firstNames", prefix), attorney.FirstNames, validate.NotEmpty()),
		validate.WithSource(fmt.Sprintf("%s/lastName", prefix), attorney.LastName, validate.NotEmpty()),
		validate.WithSource(fmt.Sprintf("%s/dateOfBirth", prefix), attorney.DateOfBirth, validate.Date()),
		validateAddress(fmt.Sprintf("%s/address", prefix), attorney.Address),
		validate.WithSource(fmt.Sprintf("%s/status", prefix), attorney.Status, validate.Valid()),
		validate.WithSource(fmt.Sprintf("%s/channel", prefix), attorney.Channel, validate.Valid()),
		validate.WithSource(fmt.Sprintf("%s/appointmentType", prefix), attorney.AppointmentType, validate.Valid()),
		validate.IfElse(attorney.Channel == shared.ChannelOnline,
			validate.WithSource(fmt.Sprintf("%s/email", prefix), attorney.Email, validate.NotEmpty()),
			validate.WithSource(fmt.Sprintf("%s/email", prefix), attorney.Email, validate.Empty())),
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
		validate.WithSource(fmt.Sprintf("%s/uid", prefix), trustCorporation.UID, validate.UUID()),
		validate.WithSource(fmt.Sprintf("%s/name", prefix), trustCorporation.Name, validate.NotEmpty()),
		validateAddress(fmt.Sprintf("%s/address", prefix), trustCorporation.Address),
		validate.WithSource(fmt.Sprintf("%s/status", prefix), trustCorporation.Status, validate.Valid()),
		validate.WithSource(fmt.Sprintf("%s/channel", prefix), trustCorporation.Channel, validate.Valid()),
		validate.WithSource(fmt.Sprintf("%s/appointmentType", prefix), trustCorporation.AppointmentType, validate.Valid()),
		validate.IfElse(trustCorporation.Channel == shared.ChannelOnline,
			validate.WithSource(fmt.Sprintf("%s/email", prefix), trustCorporation.Email, validate.NotEmpty()),
			validate.WithSource(fmt.Sprintf("%s/email", prefix), trustCorporation.Email, validate.Empty())),
	)
}
