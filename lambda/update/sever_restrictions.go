package main

import (
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
)

type SeverRestrictions struct {
	restrictionsAndConditions string
}

func (r SeverRestrictions) Apply(lpa *shared.Lpa) []shared.FieldError {
	lpa.RestrictionsAndConditions = r.restrictionsAndConditions
	lpa.RestrictionsAndConditionsImages = []shared.File{}

	var updatedRestrictionsAndConditions string

	if lpa.RestrictionsAndConditions == "" {
		updatedRestrictionsAndConditions = "All restrictions have been severed from the LPA"
	} else {
		updatedRestrictionsAndConditions = lpa.RestrictionsAndConditions
	}

	severRestrictionsAndConditionsNote := shared.Note{
		Type:     "SEVER_RESTRICTIONS_AND_CONDITIONS_V1",
		Datetime: time.Now().Format(time.RFC3339),
		Values: map[string]string{
			"updatedRestrictionsAndConditions": updatedRestrictionsAndConditions,
		},
	}

	lpa.AddNote(severRestrictionsAndConditionsNote)

	return nil
}

func validateSeverRestrictions(changes []shared.Change, lpa *shared.Lpa) (SeverRestrictions, []shared.FieldError) {
	var data SeverRestrictions

	if len(changes) == 0 {
		return data, []shared.FieldError{{Source: "/changes", Detail: "no changes provided"}}
	}

	data.restrictionsAndConditions = lpa.RestrictionsAndConditions

	errors := parse.Changes(changes).
		Field("/restrictionsAndConditions", &data.restrictionsAndConditions, parse.Optional()).
		Consumed()

	return data, errors
}
