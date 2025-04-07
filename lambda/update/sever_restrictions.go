package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
)

type SeverRestrictions struct {
	restrictionsAndConditions string
}

func (r SeverRestrictions) Apply(lpa *shared.Lpa) []shared.FieldError {
	lpa.RestrictionsAndConditions = r.restrictionsAndConditions

	return nil
}

func validateSeverRestrictions(changes []shared.Change, lpa *shared.Lpa) (SeverRestrictions, []shared.FieldError) {
	var data SeverRestrictions

	data.restrictionsAndConditions = lpa.RestrictionsAndConditions

	errors := parse.Changes(changes).
		Field("/restrictionsAndConditions", &data.restrictionsAndConditions, parse.Optional()).
		Consumed()

	return data, errors
}
