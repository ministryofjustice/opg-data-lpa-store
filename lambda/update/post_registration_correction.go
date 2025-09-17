package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
)

type PostRegistrationCorrection struct {
	Donor DonorPostRegistrationCorrection
}

type DonorPostRegistrationCorrection struct {
	FirstNames string
}

func (c DonorPostRegistrationCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	lpa.Donor.FirstNames = c.FirstNames

	return nil
}

func (c PostRegistrationCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	if fieldErrors := c.Donor.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	return nil
}

func validatePostRegistrationCorrection(changes []shared.Change, lpa *shared.Lpa) (PostRegistrationCorrection, []shared.FieldError) {
	var data PostRegistrationCorrection

	parser := parse.Changes(changes)
	errors := parser.Consumed()

	return data, errors
}
