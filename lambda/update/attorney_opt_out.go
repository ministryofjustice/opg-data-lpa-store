package main

import "github.com/ministryofjustice/opg-data-lpa-store/internal/shared"

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

	// TODO Check if status correct
	attorney.Status = shared.AttorneyStatusRemoved
	lpa.PutAttorney(attorney)

	switch len(lpa.Attorneys) {
	case 1, 2:
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

	if update.AuthorUID() == "" {
		return AttorneyOptOut{}, []shared.FieldError{{Source: "/update", Detail: "author UID missing from URN"}}
	}

	return AttorneyOptOut{AttorneyUID: update.AuthorUID()}, nil
}
