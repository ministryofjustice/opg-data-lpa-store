package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
)

type PostRegistrationCorrection struct {
	Donor                     DonorPostRegistrationCorrection
	Attorney                  AttorneyPostRegistrationCorrection
	CertificateProvider       CertificateProviderPostRegistrationCorrection
	AttorneyAppointmentType   AttorneyAppointmentPostRegistrationCorrection
	RestrictionsAndConditions string
}

type DonorPostRegistrationCorrection struct {
	shared.DonorCorrection
}

func (c DonorPostRegistrationCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	isDobChangeRequested := !c.DateOfBirth.IsZero() && c.DateOfBirth != lpa.Donor.DateOfBirth

	if isDobChangeRequested {
		return []shared.FieldError{{
			Source: "/donor/dateOfBirth",
			Detail: "The donor's date of birth cannot be changed once the LPA is registered",
		}}
	}

	lpa.Donor.FirstNames = c.FirstNames
	lpa.Donor.LastName = c.LastName
	lpa.Donor.OtherNamesKnownBy = c.OtherNamesKnownBy
	lpa.Donor.Address = c.Address
	lpa.Donor.Email = c.Email

	return nil
}

type CertificateProviderPostRegistrationCorrection struct {
	shared.CertificateProviderCorrection
}

func (c CertificateProviderPostRegistrationCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	lpa.CertificateProvider.FirstNames = c.FirstNames
	lpa.CertificateProvider.LastName = c.LastName
	lpa.CertificateProvider.Address = c.Address
	lpa.CertificateProvider.Email = c.Email
	lpa.CertificateProvider.Phone = c.Phone

	return nil
}

type AttorneyPostRegistrationCorrection struct {
	shared.AttorneyCorrection
}

func (c AttorneyPostRegistrationCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	if c.Index != nil {
		attorney := &lpa.Attorneys[*c.Index]
		attorney.FirstNames = c.FirstNames
		attorney.LastName = c.LastName
		attorney.DateOfBirth = c.DateOfBirth
		attorney.Address = c.Address
		attorney.Email = c.Email
		attorney.Mobile = c.Mobile
		attorney.CannotMakeJointDecisions = c.CannotMakeJointDecisions
	}

	return nil
}

type AttorneyAppointmentPostRegistrationCorrection struct {
	shared.AttorneyAppointmentTypeCorrection
}

func (c AttorneyAppointmentPostRegistrationCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	if !c.HowAttorneysMakeDecisions.Unset() {
		lpa.HowAttorneysMakeDecisions = c.HowAttorneysMakeDecisions
		lpa.HowAttorneysMakeDecisionsIsDefault = false
	}

	if !c.HowReplacementAttorneysMakeDecisions.Unset() {
		lpa.HowReplacementAttorneysMakeDecisions = c.HowReplacementAttorneysMakeDecisions
		lpa.HowReplacementAttorneysMakeDecisionsIsDefault = false
	}

	if !c.LifeSustainingTreatmentOption.Unset() {
		lpa.LifeSustainingTreatmentOption = c.LifeSustainingTreatmentOption
		lpa.LifeSustainingTreatmentOptionIsDefault = false
	}

	if !c.WhenTheLpaCanBeUsed.Unset() {
		lpa.WhenTheLpaCanBeUsed = c.WhenTheLpaCanBeUsed
		lpa.WhenTheLpaCanBeUsedIsDefault = false
	}

	lpa.HowAttorneysMakeDecisionsDetails = c.HowAttorneysMakeDecisionsDetails
	lpa.HowReplacementAttorneysStepIn = c.HowReplacementAttorneysStepIn
	lpa.HowReplacementAttorneysStepInDetails = c.HowReplacementAttorneysStepInDetails
	lpa.HowReplacementAttorneysMakeDecisionsDetails = c.HowReplacementAttorneysMakeDecisionsDetails

	return nil
}

func (c PostRegistrationCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	if fieldErrors := c.Donor.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	if fieldErrors := c.CertificateProvider.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	if fieldErrors := c.Attorney.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	if fieldErrors := c.AttorneyAppointmentType.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	lpa.RestrictionsAndConditions = c.RestrictionsAndConditions

	return nil
}

func validatePostRegistrationCorrection(changes []shared.Change, lpa *shared.Lpa) (PostRegistrationCorrection, []shared.FieldError) {
	var data PostRegistrationCorrection

	data.AttorneyAppointmentType.HowReplacementAttorneysStepIn = lpa.HowReplacementAttorneysStepIn
	data.AttorneyAppointmentType.HowReplacementAttorneysStepInDetails = lpa.HowReplacementAttorneysStepInDetails

	data.Donor.FirstNames = lpa.Donor.FirstNames
	data.Donor.LastName = lpa.Donor.LastName
	data.Donor.OtherNamesKnownBy = lpa.Donor.OtherNamesKnownBy
	data.Donor.Address = lpa.Donor.Address
	data.Donor.Email = lpa.Donor.Email

	data.CertificateProvider.FirstNames = lpa.CertificateProvider.FirstNames
	data.CertificateProvider.LastName = lpa.CertificateProvider.LastName
	data.CertificateProvider.Address = lpa.CertificateProvider.Address
	data.CertificateProvider.Email = lpa.CertificateProvider.Email
	data.CertificateProvider.Phone = lpa.CertificateProvider.Phone

	parser := parse.Changes(changes).
		Prefix("/donor", validatePostRegistrationDonor(&data.Donor), parse.Optional()).
		Prefix("/certificateProvider", validatePostRegistrationCertificateProvider(&data.CertificateProvider), parse.Optional()).
		Prefix("/attorneys", func(p *parse.Parser) []shared.FieldError {
			return p.
				EachKey(func(key string, p *parse.Parser) []shared.FieldError {
					attorneyIdx, ok := lpa.FindAttorneyIndex(key)

					if !ok || (data.Attorney.Index != nil && *data.Attorney.Index != attorneyIdx) {
						return p.OutOfRange()
					}

					data.Attorney.Index = &attorneyIdx
					data.Attorney.FirstNames = lpa.Attorneys[attorneyIdx].FirstNames
					data.Attorney.LastName = lpa.Attorneys[attorneyIdx].LastName
					data.Attorney.DateOfBirth = lpa.Attorneys[attorneyIdx].DateOfBirth
					data.Attorney.Address = lpa.Attorneys[attorneyIdx].Address
					data.Attorney.Email = lpa.Attorneys[attorneyIdx].Email
					data.Attorney.Mobile = lpa.Attorneys[attorneyIdx].Mobile

					if lpa.Attorneys[attorneyIdx].SignedAt != nil {
						data.Attorney.SignedAt = *lpa.Attorneys[attorneyIdx].SignedAt
					}

					return validatePostRegistrationAttorney(&data.Attorney, p)
				}).
				Consumed()
		}, parse.Optional())

	activeAttorneyCount, replacementAttorneyCount := shared.CountAttorneys(lpa.Attorneys, lpa.TrustCorporations)

	if activeAttorneyCount > 1 {
		parser.Field("/howAttorneysMakeDecisions", &data.AttorneyAppointmentType.HowAttorneysMakeDecisions,
			parse.Old(&lpa.HowAttorneysMakeDecisions),
			parse.Validate(validate.Valid()),
			parse.Optional())
	}

	attorneysJointlyForSomeSeverallyForOthers := data.AttorneyAppointmentType.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers ||
		data.AttorneyAppointmentType.HowAttorneysMakeDecisions.Unset() && lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers

	if attorneysJointlyForSomeSeverallyForOthers {
		parser.Field("/howAttorneysMakeDecisionsDetails", &data.AttorneyAppointmentType.HowAttorneysMakeDecisionsDetails,
			parse.Old(&lpa.HowAttorneysMakeDecisionsDetails),
			parse.Validate(validate.NotEmpty()))
	} else {
		parser.Field("/howAttorneysMakeDecisionsDetails", &data.AttorneyAppointmentType.HowAttorneysMakeDecisionsDetails,
			parse.Old(&lpa.HowAttorneysMakeDecisionsDetails),
			parse.Validate(validate.Empty()),
			parse.Optional())
	}

	attorneysJointlyAndSeverally := data.AttorneyAppointmentType.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyAndSeverally ||
		data.AttorneyAppointmentType.HowAttorneysMakeDecisions.Unset() && lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyAndSeverally

	if replacementAttorneyCount > 0 && attorneysJointlyAndSeverally {
		parser.Field("/howReplacementAttorneysStepIn", &data.AttorneyAppointmentType.HowReplacementAttorneysStepIn,
			parse.Validate(validate.Valid()),
			parse.Optional())
	}

	if data.AttorneyAppointmentType.HowReplacementAttorneysStepIn == shared.HowStepInAnotherWay {
		parser.Field("/howReplacementAttorneysStepInDetails", &data.AttorneyAppointmentType.HowReplacementAttorneysStepInDetails,
			parse.Validate(validate.NotEmpty()))
	} else {
		parser.Field("/howReplacementAttorneysStepInDetails", &data.AttorneyAppointmentType.HowReplacementAttorneysStepInDetails,
			parse.Validate(validate.Empty()),
			parse.Optional())
	}

	if replacementAttorneyCount > 1 && (data.AttorneyAppointmentType.HowReplacementAttorneysStepIn == shared.HowStepInAllCanNoLongerAct || !attorneysJointlyAndSeverally) {
		parser.Field("/howReplacementAttorneysMakeDecisions", &data.AttorneyAppointmentType.HowReplacementAttorneysMakeDecisions,
			parse.Old(&lpa.HowReplacementAttorneysMakeDecisions),
			parse.Validate(validate.Valid()),
			parse.Optional())
	}

	replacementsJointlyForSomeSeverallyForOthers := data.AttorneyAppointmentType.HowReplacementAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers ||
		data.AttorneyAppointmentType.HowReplacementAttorneysMakeDecisions.Unset() && lpa.HowReplacementAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers

	if replacementsJointlyForSomeSeverallyForOthers {
		parser.Field("/howReplacementAttorneysMakeDecisionsDetails", &data.AttorneyAppointmentType.HowReplacementAttorneysMakeDecisionsDetails,
			parse.Old(&lpa.HowReplacementAttorneysMakeDecisionsDetails),
			parse.Validate(validate.NotEmpty()))
	} else {
		parser.Field("/howReplacementAttorneysMakeDecisionsDetails", &data.AttorneyAppointmentType.HowReplacementAttorneysMakeDecisionsDetails,
			parse.Old(&lpa.HowReplacementAttorneysMakeDecisionsDetails),
			parse.Validate(validate.Empty()),
			parse.Optional())
	}

	if lpa.LpaType == shared.LpaTypePersonalWelfare {
		parser.Field("/lifeSustainingTreatmentOption", &data.AttorneyAppointmentType.LifeSustainingTreatmentOption,
			parse.Old(&lpa.LifeSustainingTreatmentOption),
			parse.Validate(validate.Valid()),
			parse.Optional())
	}

	if lpa.LpaType == shared.LpaTypePropertyAndAffairs {
		parser.Field("/whenTheLpaCanBeUsed", &data.AttorneyAppointmentType.WhenTheLpaCanBeUsed,
			parse.Old(&lpa.WhenTheLpaCanBeUsed),
			parse.Validate(validate.Valid()),
			parse.Optional())
	}

	parser.Field("/restrictionsAndConditions", &data.RestrictionsAndConditions, parse.Optional())

	errors := parser.Consumed()

	return data, errors
}

func validatePostRegistrationAttorney(attorney *AttorneyPostRegistrationCorrection, p *parse.Parser) []shared.FieldError {
	return p.
		Field("/firstNames", &attorney.FirstNames, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Field("/lastName", &attorney.LastName, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Field("/dateOfBirth", &attorney.DateOfBirth, parse.Validate(validate.Date()), parse.Optional()).
		Field("/email", &attorney.Email, parse.Optional()).
		Field("/mobile", &attorney.Mobile, parse.Optional()).
		Prefix("/address", validatePostRegistrationAddress(&attorney.Address), parse.Optional()).
		Field(signedAt, &attorney.SignedAt, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Field("/cannotMakeJointDecisions", &attorney.CannotMakeJointDecisions, parse.Optional()).
		Consumed()
}

func validatePostRegistrationDonor(donor *DonorPostRegistrationCorrection) func(p *parse.Parser) []shared.FieldError {
	return func(p *parse.Parser) []shared.FieldError {
		return p.
			Field("/firstNames", &donor.FirstNames, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Field("/lastName", &donor.LastName, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Field("/otherNamesKnownBy", &donor.OtherNamesKnownBy, parse.Optional()).
			Field("/dateOfBirth", &donor.DateOfBirth, parse.Validate(validate.Date()), parse.Optional()).
			Prefix("/address", validatePostRegistrationAddress(&donor.Address), parse.Optional()).
			Field("/email", &donor.Email, parse.Optional()).
			Consumed()
	}
}

func validatePostRegistrationCertificateProvider(certificateProvider *CertificateProviderPostRegistrationCorrection) func(p *parse.Parser) []shared.FieldError {
	return func(p *parse.Parser) []shared.FieldError {
		return p.
			Field("/firstNames", &certificateProvider.FirstNames, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Field("/lastName", &certificateProvider.LastName, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Prefix("/address", validatePostRegistrationAddress(&certificateProvider.Address), parse.Optional()).
			Field("/email", &certificateProvider.Email, parse.Optional()).
			Field("/phone", &certificateProvider.Phone, parse.Optional()).
			Field("/signedAt", &certificateProvider.SignedAt, parse.Optional()).
			Consumed()
	}
}

func validatePostRegistrationAddress(address *shared.Address) func(p *parse.Parser) []shared.FieldError {
	return func(p *parse.Parser) []shared.FieldError {
		return p.
			Field("/line1", &address.Line1, parse.Optional()).
			Field("/line2", &address.Line2, parse.Optional()).
			Field("/line3", &address.Line3, parse.Optional()).
			Field("/town", &address.Town, parse.Optional()).
			Field("/postcode", &address.Postcode, parse.Optional()).
			Field("/country", &address.Country, parse.Validate(validate.Country()), parse.Optional()).
			Consumed()
	}
}
