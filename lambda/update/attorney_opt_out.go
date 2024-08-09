package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
)

type AttorneyOptOut struct {
	AttorneyUID string
}

func (c AttorneyOptOut) Apply(lpa *shared.Lpa) []shared.FieldError {
	attorney, ok := lpa.GetAttorney(c.AttorneyUID)
	if !ok {
		return []shared.FieldError{{Source: "/type", Detail: "attorney not found"}}
	}

	if attorney.SignedAt != nil && !attorney.SignedAt.IsZero() {
		return []shared.FieldError{{Source: "/type", Detail: "attorney cannot opt out after signing"}}
	}

	attorney.Status = shared.AttorneyStatusRemoved
	lpa.PutAttorney(attorney)

	var activeAttorneys []any
	activeAttorneys = append(activeAttorneys, lpa.ActiveAttorneys())
	activeAttorneys = append(activeAttorneys, lpa.ActiveTrustCorporations())

	switch len(activeAttorneys) {
	case 0, 1:
		lpa.Status = shared.LpaStatusCannotRegister
	default:
		if lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointly || lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers {
			lpa.Status = shared.LpaStatusCannotRegister
		}
	}

	return nil
}

func validateAttorneyOptOut(update shared.Update) (AttorneyOptOut, []shared.FieldError) {
	if len(update.Changes) > 0 {
		return AttorneyOptOut{}, []shared.FieldError{{Source: "/changes", Detail: "expected empty"}}
	}

	author := update.Author.Details()

	uidErrors := validate.All(
		validate.UUID("/update/author/uid", author.UID),
		validate.UUID("/update/subject", update.Subject),
	)

	if len(uidErrors) > 0 {
		return AttorneyOptOut{}, uidErrors
	}

	if author.Service == "makeregister" && update.Subject != author.UID {
		return AttorneyOptOut{}, []shared.FieldError{{Source: "/update", Detail: "cannot change other actors"}}
	}

	return AttorneyOptOut{AttorneyUID: update.Subject}, nil
}
