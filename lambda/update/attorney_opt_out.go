package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
)

type AttorneyOptOut struct {
	AttorneyUID string
}

func (c AttorneyOptOut) Apply(lpa *shared.Lpa) []shared.FieldError {
	for i := range lpa.Attorneys {
		if lpa.Attorneys[i].UID == c.AttorneyUID {
			if lpa.Attorneys[i].SignedAt != nil && !lpa.Attorneys[i].SignedAt.IsZero() {
				return []shared.FieldError{{Source: "/type", Detail: "attorney cannot opt out after signing"}}
			}

			lpa.Attorneys[i].Status = shared.AttorneyStatusRemoved
			return nil
		}
	}

	return []shared.FieldError{{Source: "/type", Detail: "attorney not found"}}
}

func validateAttorneyOptOut(update shared.Update) (AttorneyOptOut, []shared.FieldError) {
	if len(update.Changes) > 0 {
		return AttorneyOptOut{}, []shared.FieldError{{Source: "/changes", Detail: "expected empty"}}
	}

	author := update.Author.Details()

	if errs := validate.WithSource("/author", author.UID, validate.UUID()); len(errs) > 0 {
		return AttorneyOptOut{}, errs
	}

	return AttorneyOptOut{AttorneyUID: author.UID}, nil
}
