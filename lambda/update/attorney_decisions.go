package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
	"strconv"
)

type AttorneyDecision struct {
	ChangeAttorneyDecisions []ChangeAttorneyDecisions
}

type ChangeAttorneyDecisions struct {
	Index     *int
	Decisions bool
}

func (a AttorneyDecision) Apply(lpa *shared.Lpa) []shared.FieldError {
	for _, changeAttorneyDecisions := range a.ChangeAttorneyDecisions {
		source := "/attorneys/" + strconv.Itoa(*changeAttorneyDecisions.Index) + "/cannotMakeJointDecisions"

		if lpa.HowAttorneysMakeDecisions != shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers && lpa.Attorneys[*changeAttorneyDecisions.Index].AppointmentType == shared.AppointmentTypeOriginal {
			return []shared.FieldError{{Source: source, Detail: "The appointment type must be jointly for some and severally for others"}}
		}

		if lpa.HowReplacementAttorneysMakeDecisions != shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers && lpa.Attorneys[*changeAttorneyDecisions.Index].AppointmentType == shared.AppointmentTypeReplacement {
			return []shared.FieldError{{Source: source, Detail: "The appointment type must be jointly for some and severally for others"}}
		}

		lpa.Attorneys[*changeAttorneyDecisions.Index].CannotMakeJointDecisions = changeAttorneyDecisions.Decisions
	}

	return nil
}

func validateAttorneyDecisions(changes []shared.Change, lpa *shared.Lpa) (AttorneyDecision, []shared.FieldError) {
	var data AttorneyDecision
	key := -1

	errors := parse.Changes(changes).
		Prefix("/attorneys", func(p *parse.Parser) []shared.FieldError {
			return p.
				Each(func(i int, p *parse.Parser) []shared.FieldError {
					key++
					data.ChangeAttorneyDecisions = append(data.ChangeAttorneyDecisions, ChangeAttorneyDecisions{Index: &i, Decisions: lpa.Attorneys[i].CannotMakeJointDecisions})

					return p.
						Field("/cannotMakeJointDecisions", &data.ChangeAttorneyDecisions[key].Decisions).
						Consumed()
				}).
				Consumed()
		}).
		Consumed()

	return data, errors
}
