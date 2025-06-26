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

		if (lpa.HowAttorneysMakeDecisions != shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers && lpa.Attorneys[*changeAttorneyDecisions.Index].AppointmentType == shared.AppointmentTypeOriginal) ||
			(lpa.HowReplacementAttorneysMakeDecisions != shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers && lpa.Attorneys[*changeAttorneyDecisions.Index].AppointmentType == shared.AppointmentTypeReplacement) {
			return []shared.FieldError{{Source: source, Detail: "The appointment type must be jointly for some and severally for others"}}
		}

		lpa.Attorneys[*changeAttorneyDecisions.Index].CannotMakeJointDecisions = changeAttorneyDecisions.Decisions
	}

	return nil
}

func validateAttorneyDecisions(changes []shared.Change, lpa *shared.Lpa) (AttorneyDecision, []shared.FieldError) {
	var data AttorneyDecision
	i := -1

	errors := parse.Changes(changes).
		Prefix("/attorneys", func(p *parse.Parser) []shared.FieldError {
			return p.
				EachKey(func(key string, p *parse.Parser) []shared.FieldError {
					i++

					attorneyIdx, ok := lpa.FindAttorneyIndex(key)
					if !ok {
						return p.OutOfRange()
					}

					data.ChangeAttorneyDecisions = append(data.ChangeAttorneyDecisions, ChangeAttorneyDecisions{Index: &attorneyIdx, Decisions: lpa.Attorneys[attorneyIdx].CannotMakeJointDecisions})

					return p.
						Field("/cannotMakeJointDecisions", &data.ChangeAttorneyDecisions[i].Decisions).
						Consumed()
				}).
				Consumed()
		}).
		Consumed()

	return data, errors
}
