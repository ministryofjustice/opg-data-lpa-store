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

	return nil
}

func validateAttorneyOptOut(update shared.Update) (AttorneyOptOut, []shared.FieldError) {
	if len(update.Changes) > 0 {
		return AttorneyOptOut{}, []shared.FieldError{{Source: "/changes", Detail: "expected empty"}}
	}

	author := update.Author.Details()

	if errs := validate.UUID("/author", author.UID); len(errs) > 0 {
		return AttorneyOptOut{}, errs
	}

	return AttorneyOptOut{AttorneyUID: author.UID}, nil
}
