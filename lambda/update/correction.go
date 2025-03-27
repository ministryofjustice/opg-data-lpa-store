package main

import (
	"strconv"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
)

const signedAt = "/signedAt"

type Correction struct {
	Donor                                       DonorCorrection
	Attorney                                    AttorneyCorrection
	CertificateProvider                         CertificateProviderCorrection
	HowAttorneysMakeDecisions                   shared.HowMakeDecisions
	HowAttorneysMakeDecisionsDetails            string
	HowReplacementAttorneysStepIn               shared.HowStepIn
	HowReplacementAttorneysStepInDetails        string
	HowReplacementAttorneysMakeDecisions        shared.HowMakeDecisions
	HowReplacementAttorneysMakeDecisionsDetails string
	LifeSustainingTreatmentOption               shared.LifeSustainingTreatment
	WhenTheLpaCanBeUsed                         shared.CanUse
	SignedAt                                    time.Time
}

type DonorCorrection struct {
	FirstNames        string
	LastName          string
	OtherNamesKnownBy string
	DateOfBirth       shared.Date
	Address           shared.Address
	Email             string
}

func (c DonorCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	lpa.Donor.FirstNames = c.FirstNames
	lpa.Donor.LastName = c.LastName
	lpa.Donor.OtherNamesKnownBy = c.OtherNamesKnownBy
	lpa.Donor.DateOfBirth = c.DateOfBirth
	lpa.Donor.Address = c.Address
	lpa.Donor.Email = c.Email

	return nil
}

type CertificateProviderCorrection struct {
	FirstNames string
	LastName   string
	Address    shared.Address
	Email      string
	Phone      string
	SignedAt   time.Time
}

func (c CertificateProviderCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	if !c.SignedAt.IsZero() && !c.SignedAt.Equal(*lpa.CertificateProvider.SignedAt) && lpa.Channel == shared.ChannelOnline {
		return []shared.FieldError{{
			Source: "/certificateProvider" + signedAt,
			Detail: "The Certificate Provider Signed on date cannot be changed for online LPAs",
		}}
	}

	lpa.CertificateProvider.FirstNames = c.FirstNames
	lpa.CertificateProvider.LastName = c.LastName
	lpa.CertificateProvider.Address = c.Address
	lpa.CertificateProvider.Email = c.Email
	lpa.CertificateProvider.Phone = c.Phone
	lpa.CertificateProvider.SignedAt = &c.SignedAt

	return nil
}

type AttorneyCorrection struct {
	Index       *int
	FirstNames  string
	LastName    string
	DateOfBirth shared.Date
	Address     shared.Address
	Email       string
	Mobile      string
	SignedAt    time.Time
}

func (c AttorneyCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	if c.Index != nil {
		if !c.SignedAt.IsZero() && !c.SignedAt.Equal(*lpa.Attorneys[*c.Index].SignedAt) && lpa.Channel == shared.ChannelOnline {
			source := "/attorney/" + strconv.Itoa(*c.Index) + signedAt
			return []shared.FieldError{{Source: source, Detail: "The attorney signed at date cannot be changed for online LPA"}}
		}

		attorney := &lpa.Attorneys[*c.Index]
		attorney.FirstNames = c.FirstNames
		attorney.LastName = c.LastName
		attorney.DateOfBirth = c.DateOfBirth
		attorney.Address = c.Address
		attorney.Email = c.Email
		attorney.Mobile = c.Mobile
		attorney.SignedAt = &c.SignedAt
	}

	return nil
}

func (c Correction) Apply(lpa *shared.Lpa) []shared.FieldError {
	if !c.SignedAt.IsZero() && !c.SignedAt.Equal(lpa.SignedAt) && lpa.Channel == shared.ChannelOnline {
		return []shared.FieldError{{Source: signedAt, Detail: "LPA Signed on date cannot be changed for online LPAs"}}
	}

	if lpa.Status == shared.LpaStatusRegistered {
		return []shared.FieldError{{Source: "/type", Detail: "Cannot make corrections to a Registered LPA"}}
	}

	if fieldErrors := c.Donor.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	if fieldErrors := c.CertificateProvider.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	if fieldErrors := c.Attorney.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	lpa.SignedAt = c.SignedAt

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

func validateCorrection(changes []shared.Change, lpa *shared.Lpa) (Correction, []shared.FieldError) {
	var data Correction

	data.SignedAt = lpa.SignedAt
	data.HowReplacementAttorneysStepIn = lpa.HowReplacementAttorneysStepIn
	data.HowReplacementAttorneysStepInDetails = lpa.HowReplacementAttorneysStepInDetails

	data.Donor.FirstNames = lpa.Donor.FirstNames
	data.Donor.LastName = lpa.Donor.LastName
	data.Donor.OtherNamesKnownBy = lpa.Donor.OtherNamesKnownBy
	data.Donor.DateOfBirth = lpa.Donor.DateOfBirth
	data.Donor.Address = lpa.Donor.Address
	data.Donor.Email = lpa.Donor.Email

	data.CertificateProvider.FirstNames = lpa.CertificateProvider.FirstNames
	data.CertificateProvider.LastName = lpa.CertificateProvider.LastName
	data.CertificateProvider.Address = lpa.CertificateProvider.Address
	data.CertificateProvider.Email = lpa.CertificateProvider.Email
	data.CertificateProvider.Phone = lpa.CertificateProvider.Phone
	if lpa.CertificateProvider.SignedAt != nil {
		data.CertificateProvider.SignedAt = *lpa.CertificateProvider.SignedAt
	}

	parser := parse.Changes(changes).
		Field(signedAt, &data.SignedAt, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Prefix("/donor", validateDonor(&data.Donor), parse.Optional()).
		Prefix("/certificateProvider", validateCertificateProvider(&data.CertificateProvider), parse.Optional()).
		Prefix("/attorneys", func(p *parse.Parser) []shared.FieldError {
			return p.
				Each(func(i int, p *parse.Parser) []shared.FieldError {
					if data.Attorney.Index != nil && *data.Attorney.Index != i {
						return p.OutOfRange()
					}

					data.Attorney.Index = &i
					data.Attorney.FirstNames = lpa.Attorneys[i].FirstNames
					data.Attorney.LastName = lpa.Attorneys[i].LastName
					data.Attorney.DateOfBirth = lpa.Attorneys[i].DateOfBirth
					data.Attorney.Address = lpa.Attorneys[i].Address
					data.Attorney.Email = lpa.Attorneys[i].Email
					data.Attorney.Mobile = lpa.Attorneys[i].Mobile

					if lpa.Attorneys[i].SignedAt != nil {
						data.Attorney.SignedAt = *lpa.Attorneys[i].SignedAt
					}

					return validateAttorney(&data.Attorney, p)
				}).
				Consumed()
		}, parse.Optional())

	activeAttorneyCount, replacementAttorneyCount := shared.CountAttorneys(lpa.Attorneys, lpa.TrustCorporations)

	if activeAttorneyCount > 1 {
		parser.Field("/howAttorneysMakeDecisions", &data.HowAttorneysMakeDecisions,
			parse.Old(&lpa.HowAttorneysMakeDecisions),
			parse.Validate(validate.Valid()),
			parse.Optional())
	}

	attorneysJointlyForSomeSeverallyForOthers := data.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers ||
		data.HowAttorneysMakeDecisions.Unset() && lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers

	if attorneysJointlyForSomeSeverallyForOthers {
		parser.Field("/howAttorneysMakeDecisionsDetails", &data.HowAttorneysMakeDecisionsDetails,
			parse.Old(&lpa.HowAttorneysMakeDecisionsDetails),
			parse.Validate(validate.NotEmpty()))
	} else {
		parser.Field("/howAttorneysMakeDecisionsDetails", &data.HowAttorneysMakeDecisionsDetails,
			parse.Old(&lpa.HowAttorneysMakeDecisionsDetails),
			parse.Validate(validate.Empty()),
			parse.Optional())
	}

	attorneysJointlyAndSeverally := data.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyAndSeverally ||
		data.HowAttorneysMakeDecisions.Unset() && lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyAndSeverally

	if replacementAttorneyCount > 0 && attorneysJointlyAndSeverally {
		parser.Field("/howReplacementAttorneysStepIn", &data.HowReplacementAttorneysStepIn,
			parse.Validate(validate.Valid()),
			parse.Optional())
	}

	if data.HowReplacementAttorneysStepIn == shared.HowStepInAnotherWay {
		parser.Field("/howReplacementAttorneysStepInDetails", &data.HowReplacementAttorneysStepInDetails,
			parse.Validate(validate.NotEmpty()))
	} else {
		parser.Field("/howReplacementAttorneysStepInDetails", &data.HowReplacementAttorneysStepInDetails,
			parse.Validate(validate.Empty()),
			parse.Optional())
	}

	if replacementAttorneyCount > 1 && (data.HowReplacementAttorneysStepIn == shared.HowStepInAllCanNoLongerAct || !attorneysJointlyAndSeverally) {
		parser.Field("/howReplacementAttorneysMakeDecisions", &data.HowReplacementAttorneysMakeDecisions,
			parse.Old(&lpa.HowReplacementAttorneysMakeDecisions),
			parse.Validate(validate.Valid()),
			parse.Optional())
	}

	replacementsJointlyForSomeSeverallyForOthers := data.HowReplacementAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers ||
		data.HowReplacementAttorneysMakeDecisions.Unset() && lpa.HowReplacementAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers

	if replacementsJointlyForSomeSeverallyForOthers {
		parser.Field("/howReplacementAttorneysMakeDecisionsDetails", &data.HowReplacementAttorneysMakeDecisionsDetails,
			parse.Old(&lpa.HowReplacementAttorneysMakeDecisionsDetails),
			parse.Validate(validate.NotEmpty()))
	} else {
		parser.Field("/howReplacementAttorneysMakeDecisionsDetails", &data.HowReplacementAttorneysMakeDecisionsDetails,
			parse.Old(&lpa.HowReplacementAttorneysMakeDecisionsDetails),
			parse.Validate(validate.Empty()),
			parse.Optional())
	}

	if lpa.LpaType == shared.LpaTypePersonalWelfare {
		parser.Field("/lifeSustainingTreatmentOption", &data.LifeSustainingTreatmentOption,
			parse.Old(&lpa.LifeSustainingTreatmentOption),
			parse.Validate(validate.Valid()),
			parse.Optional())
	}

	if lpa.LpaType == shared.LpaTypePropertyAndAffairs {
		parser.Field("/whenTheLpaCanBeUsed", &data.WhenTheLpaCanBeUsed,
			parse.Old(&lpa.WhenTheLpaCanBeUsed),
			parse.Validate(validate.Valid()),
			parse.Optional())
	}

	errors := parser.Consumed()

	return data, errors
}

func validateAttorney(attorney *AttorneyCorrection, p *parse.Parser) []shared.FieldError {
	return p.
		Field("/firstNames", &attorney.FirstNames, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Field("/lastName", &attorney.LastName, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Field("/dateOfBirth", &attorney.DateOfBirth, parse.Validate(validate.Date()), parse.Optional()).
		Field("/email", &attorney.Email, parse.Optional()).
		Field("/mobile", &attorney.Mobile, parse.Optional()).
		Prefix("/address", validateAddress(&attorney.Address), parse.Optional()).
		Field(signedAt, &attorney.SignedAt, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Consumed()
}

func validateDonor(donor *DonorCorrection) func(p *parse.Parser) []shared.FieldError {
	return func(p *parse.Parser) []shared.FieldError {
		return p.
			Field("/firstNames", &donor.FirstNames, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Field("/lastName", &donor.LastName, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Field("/otherNamesKnownBy", &donor.OtherNamesKnownBy, parse.Optional()).
			Field("/dateOfBirth", &donor.DateOfBirth, parse.Validate(validate.Date()), parse.Optional()).
			Prefix("/address", validateAddress(&donor.Address), parse.Optional()).
			Field("/email", &donor.Email, parse.Optional()).
			Consumed()
	}
}

func validateCertificateProvider(certificateProvider *CertificateProviderCorrection) func(p *parse.Parser) []shared.FieldError {
	return func(p *parse.Parser) []shared.FieldError {
		return p.
			Field("/firstNames", &certificateProvider.FirstNames, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Field("/lastName", &certificateProvider.LastName, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Prefix("/address", validateAddress(&certificateProvider.Address), parse.Optional()).
			Field("/email", &certificateProvider.Email, parse.Optional()).
			Field("/phone", &certificateProvider.Phone, parse.Optional()).
			Field("/signedAt", &certificateProvider.SignedAt, parse.Optional()).
			Consumed()
	}
}

func validateAddress(address *shared.Address) func(p *parse.Parser) []shared.FieldError {
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
