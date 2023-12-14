package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

func Validate(lpa shared.LpaInit) []shared.FieldError {
	activeAttorneyCount, replacementAttorneyCount := countAttorneys(lpa.Attorneys)

	return flatten(
		validateIsValid("/type", lpa.Type),
		required("/donor/firstNames", lpa.Donor.FirstNames),
		required("/donor/lastName", lpa.Donor.LastName),
		validateDate("/donor/dateOfBirth", lpa.Donor.DateOfBirth),
		validateAddress("/donor/address", lpa.Donor.Address),
		required("/certificateProvider/firstNames", lpa.CertificateProvider.FirstNames),
		required("/certificateProvider/lastName", lpa.CertificateProvider.LastName),
		validateAddress("/certificateProvider/address", lpa.CertificateProvider.Address),
		validateIsValid("/certificateProvider/carryOutBy", lpa.CertificateProvider.CarryOutBy),
		validateIfElse(lpa.CertificateProvider.CarryOutBy == shared.CarryOutByOnline,
			required("/certificateProvider/email", lpa.CertificateProvider.Email),
			empty("/certificateProvider/email", lpa.CertificateProvider.Email)),
		validateAttorneys("/attorneys", lpa.Attorneys),
		validateIfElse(activeAttorneyCount > 1,
			validateIsValid("/howAttorneysMakeDecisions", lpa.HowAttorneysMakeDecisions),
			validateUnset("/howAttorneysMakeDecisions", lpa.HowAttorneysMakeDecisions)),
		validateIfElse(lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
			required("/howAttorneysMakeDecisionsDetails", lpa.HowAttorneysMakeDecisionsDetails),
			empty("/howAttorneysMakeDecisionsDetails", lpa.HowAttorneysMakeDecisionsDetails)),
		validateIf(lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyAndSeverally,
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
		validateIf(lpa.Type == "hw", flatten(
			validateIsValid("/lifeSustainingTreatmentOption", lpa.LifeSustainingTreatmentOption),
			validateUnset("/whenTheLpaCanBeUsed", lpa.WhenTheLpaCanBeUsed))),
		validateIf(lpa.Type == "pfa", flatten(
			validateIsValid("/whenTheLpaCanBeUsed", lpa.WhenTheLpaCanBeUsed),
			validateUnset("/lifeSustainingTreatmentOption", lpa.LifeSustainingTreatmentOption))),
		validateTime("/signedAt", lpa.SignedAt),
	)
}

func countAttorneys(as []shared.Attorney) (actives, replacements int) {
	for _, a := range as {
		switch a.Status {
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

func validateIf(ok bool, e []shared.FieldError) []shared.FieldError {
	if ok {
		return e
	}

	return nil
}

func validateIfElse(ok bool, eIf []shared.FieldError, eElse []shared.FieldError) []shared.FieldError {
	if ok {
		return eIf
	}

	return eElse
}

func required(source string, value string) []shared.FieldError {
	if value == "" {
		return []shared.FieldError{{Source: source, Detail: "field is required"}}
	}

	return nil
}

func empty(source string, value string) []shared.FieldError {
	if value != "" {
		return []shared.FieldError{{Source: source, Detail: "field must not be provided"}}
	}

	return nil
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
	if t.IsZero() {
		return []shared.FieldError{{Source: source, Detail: "field is required"}}
	}

	return nil
}

func validateAddress(prefix string, address shared.Address) []shared.FieldError {
	errors := flatten(
		required(fmt.Sprintf("%s/line1", prefix), address.Line1),
		required(fmt.Sprintf("%s/town", prefix), address.Town),
		required(fmt.Sprintf("%s/country", prefix), address.Country),
	)

	if ok, _ := regexp.MatchString("^[A-Z]{2}$", address.Country); !ok {
		errors = append(errors, shared.FieldError{Source: fmt.Sprintf("%s/country", prefix), Detail: "must be a valid ISO-3166-1 country code"})
	}

	return errors
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
	if !v.Unset() {
		return []shared.FieldError{{Source: source, Detail: "field must not be provided"}}
	}

	return nil
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
