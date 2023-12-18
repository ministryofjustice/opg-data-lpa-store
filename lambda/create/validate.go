package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

var countryCodeRe = regexp.MustCompile("^[A-Z]{2}$")

func Validate(lpa shared.LpaInit) []shared.FieldError {
	activeAttorneyCount, replacementAttorneyCount := countAttorneys(lpa.Attorneys, lpa.TrustCorporations)

	return flatten(
		validateIsValid("/lpaType", lpa.LpaType),
		required("/donor/firstNames", lpa.Donor.FirstNames),
		required("/donor/lastName", lpa.Donor.LastName),
		validateDate("/donor/dateOfBirth", lpa.Donor.DateOfBirth),
		validateAddress("/donor/address", lpa.Donor.Address),
		required("/certificateProvider/firstNames", lpa.CertificateProvider.FirstNames),
		required("/certificateProvider/lastName", lpa.CertificateProvider.LastName),
		validateAddress("/certificateProvider/address", lpa.CertificateProvider.Address),
		validateIsValid("/certificateProvider/channel", lpa.CertificateProvider.Channel),
		validateIfElse(lpa.CertificateProvider.Channel == shared.ChannelOnline,
			required("/certificateProvider/email", lpa.CertificateProvider.Email),
			empty("/certificateProvider/email", lpa.CertificateProvider.Email)),
		validateAttorneys("/attorneys", lpa.Attorneys),
		validateTrustCorporations("/trustCorporations", lpa.TrustCorporations),
		validateIfElse(activeAttorneyCount > 1,
			validateIsValid("/howAttorneysMakeDecisions", lpa.HowAttorneysMakeDecisions),
			validateUnset("/howAttorneysMakeDecisions", lpa.HowAttorneysMakeDecisions)),
		validateIfElse(lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
			required("/howAttorneysMakeDecisionsDetails", lpa.HowAttorneysMakeDecisionsDetails),
			empty("/howAttorneysMakeDecisionsDetails", lpa.HowAttorneysMakeDecisionsDetails)),
		validateIf(replacementAttorneyCount > 0 && lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyAndSeverally,
			validateIsValid("/howReplacementAttorneysStepIn", lpa.HowReplacementAttorneysStepIn)),
		validateIfElse(lpa.HowReplacementAttorneysStepIn == shared.HowStepInAnotherWay,
			required("/howReplacementAttorneysStepInDetails", lpa.HowReplacementAttorneysStepInDetails),
			empty("/howReplacementAttorneysStepInDetails", lpa.HowReplacementAttorneysStepInDetails)),
		validateIfElse(replacementAttorneyCount > 1 && (lpa.HowReplacementAttorneysStepIn == shared.HowStepInAllCanNoLongerAct || lpa.HowAttorneysMakeDecisions != shared.HowMakeDecisionsJointlyAndSeverally),
			validateIsValid("/howReplacementAttorneysMakeDecisions", lpa.HowReplacementAttorneysMakeDecisions),
			validateUnset("/howReplacementAttorneysMakeDecisions", lpa.HowReplacementAttorneysMakeDecisions)),
		validateIfElse(lpa.HowReplacementAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
			required("/howReplacementAttorneysMakeDecisionsDetails", lpa.HowReplacementAttorneysMakeDecisionsDetails),
			empty("/howReplacementAttorneysMakeDecisionsDetails", lpa.HowReplacementAttorneysMakeDecisionsDetails)),
		validateIf(lpa.LpaType == shared.LpaTypePersonalWelfare, flatten(
			validateIsValid("/lifeSustainingTreatmentOption", lpa.LifeSustainingTreatmentOption),
			validateUnset("/whenTheLpaCanBeUsed", lpa.WhenTheLpaCanBeUsed))),
		validateIf(lpa.LpaType == shared.LpaTypePropertyAndAffairs, flatten(
			validateIsValid("/whenTheLpaCanBeUsed", lpa.WhenTheLpaCanBeUsed),
			validateUnset("/lifeSustainingTreatmentOption", lpa.LifeSustainingTreatmentOption))),
		validateTime("/signedAt", lpa.SignedAt),
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

func flatten(fieldErrors ...[]shared.FieldError) []shared.FieldError {
	var errors []shared.FieldError

	for _, e := range fieldErrors {
		if e != nil {
			errors = append(errors, e...)
		}
	}

	return errors
}

func validateIfElse(ok bool, eIf []shared.FieldError, eElse []shared.FieldError) []shared.FieldError {
	if ok {
		return eIf
	}

	return eElse
}

func validateIf(ok bool, e []shared.FieldError) []shared.FieldError {
	return validateIfElse(ok, e, nil)
}

func required(source string, value string) []shared.FieldError {
	return validateIf(value == "", []shared.FieldError{{Source: source, Detail: "field is required"}})
}

func empty(source string, value string) []shared.FieldError {
	return validateIf(value != "", []shared.FieldError{{Source: source, Detail: "field must not be provided"}})
}

func validateDate(source string, date shared.Date) []shared.FieldError {
	if date.IsMalformed {
		return []shared.FieldError{{Source: source, Detail: "invalid format"}}
	}

	if date.IsZero() {
		return []shared.FieldError{{Source: source, Detail: "field is required"}}
	}

	return nil
}

func validateTime(source string, t time.Time) []shared.FieldError {
	return validateIf(t.IsZero(), []shared.FieldError{{Source: source, Detail: "field is required"}})
}

func validateAddress(prefix string, address shared.Address) []shared.FieldError {
	return flatten(
		required(fmt.Sprintf("%s/line1", prefix), address.Line1),
		required(fmt.Sprintf("%s/town", prefix), address.Town),
		required(fmt.Sprintf("%s/country", prefix), address.Country),
		validateIf(!countryCodeRe.MatchString(address.Country), []shared.FieldError{{Source: fmt.Sprintf("%s/country", prefix), Detail: "must be a valid ISO-3166-1 country code"}}),
	)
}

type isValid interface {
	~string
	IsValid() bool
}

func validateIsValid[V isValid](source string, v V) []shared.FieldError {
	if e := required(source, string(v)); e != nil {
		return e
	}

	if !v.IsValid() {
		return []shared.FieldError{{Source: source, Detail: "invalid value"}}
	}

	return nil
}

func validateUnset(source string, v interface{ Unset() bool }) []shared.FieldError {
	return validateIf(!v.Unset(), []shared.FieldError{{Source: source, Detail: "field must not be provided"}})
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
	return flatten(
		required(fmt.Sprintf("%s/firstNames", prefix), attorney.FirstNames),
		required(fmt.Sprintf("%s/lastName", prefix), attorney.LastName),
		required(fmt.Sprintf("%s/status", prefix), string(attorney.Status)),
		validateDate(fmt.Sprintf("%s/dateOfBirth", prefix), attorney.DateOfBirth),
		validateAddress(fmt.Sprintf("%s/address", prefix), attorney.Address),
		validateIsValid(fmt.Sprintf("%s/status", prefix), attorney.Status),
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
	return flatten(
		required(fmt.Sprintf("%s/name", prefix), trustCorporation.Name),
		required(fmt.Sprintf("%s/companyNumber", prefix), trustCorporation.CompanyNumber),
		required(fmt.Sprintf("%s/email", prefix), trustCorporation.Email),
		validateAddress(fmt.Sprintf("%s/address", prefix), trustCorporation.Address),
		validateIsValid(fmt.Sprintf("%s/status", prefix), trustCorporation.Status),
	)
}
