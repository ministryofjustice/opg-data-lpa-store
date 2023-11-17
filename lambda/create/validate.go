package main

import (
	"fmt"
	"regexp"

	"github.com/ministryofjustice/opg-data-lpa-store/lambda/shared"
)

func validateAddress(address shared.Address, prefix string, errors []shared.FieldError) []shared.FieldError {
	requiredFields := map[string]string{
		fmt.Sprintf("%s/line1", prefix):   address.Line1,
		fmt.Sprintf("%s/town", prefix):    address.Town,
		fmt.Sprintf("%s/country", prefix): address.Country,
	}

	for source, value := range requiredFields {
		if value == "" {
			errors = append(errors, shared.FieldError{Source: source, Detail: "field is required"})
		}
	}

	if ok, _ := regexp.MatchString("^[A-Z]{2}$", address.Country); !ok {
		errors = append(errors, shared.FieldError{Source: fmt.Sprintf("%s/country", prefix), Detail: "must be a valid ISO-3166-1 country code"})
	}

	return errors
}

func validateAttorney(attorney shared.Attorney, prefix string, errors []shared.FieldError) []shared.FieldError {
	requiredFields := map[string]string{
		fmt.Sprintf("%s/firstNames", prefix): attorney.FirstNames,
		fmt.Sprintf("%s/surname", prefix):    attorney.Surname,
		fmt.Sprintf("%s/status", prefix):     string(attorney.Status),
	}

	for source, value := range requiredFields {
		if value == "" {
			errors = append(errors, shared.FieldError{Source: source, Detail: "field is required"})
		}
	}

	if attorney.DateOfBirth.IsZero() {
		errors = append(errors, shared.FieldError{Source: fmt.Sprintf("%s/dateOfBirth", prefix), Detail: "field is required"})
	}

	if !attorney.Status.IsValid() {
		errors = append(errors, shared.FieldError{Source: fmt.Sprintf("%s/status", prefix), Detail: "invalid value"})
	}

	errors = validateAddress(attorney.Address, fmt.Sprintf("%s/address", prefix), errors)

	return errors
}

func Validate(lpa shared.LpaInit) []shared.FieldError {
	errors := []shared.FieldError{}

	requiredFields := map[string]string{
		"/donor/firstNames": lpa.Donor.FirstNames,
		"/donor/surname":    lpa.Donor.Surname,
	}

	for source, value := range requiredFields {
		if value == "" {
			errors = append(errors, shared.FieldError{Source: source, Detail: "field is required"})
		}
	}

	if lpa.Donor.DateOfBirth.IsMalformed {
		errors = append(errors, shared.FieldError{Source: "/donor/dateOfBirth", Detail: "invalid format"})
	} else if lpa.Donor.DateOfBirth.IsZero() {
		errors = append(errors, shared.FieldError{Source: "/donor/dateOfBirth", Detail: "field is required"})
	}

	errors = validateAddress(lpa.Donor.Address, "/donor/address", errors)

	if len(lpa.Attorneys) == 0 {
		errors = append(errors, shared.FieldError{Source: "/attorneys", Detail: "at least one attorney is required"})
	}

	for index, attorney := range lpa.Attorneys {
		prefix := fmt.Sprintf("/attorneys/%d", index)
		errors = validateAttorney(attorney, prefix, errors)
	}

	return errors
}
