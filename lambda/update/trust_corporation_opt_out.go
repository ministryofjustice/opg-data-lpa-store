package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
)

type TrustCorporationOptOut struct {
	trustCorporationUID string
}

func (c TrustCorporationOptOut) Apply(lpa *shared.Lpa) []shared.FieldError {
	for i := range lpa.TrustCorporations {
		if lpa.TrustCorporations[i].UID == c.trustCorporationUID {
			if len(lpa.TrustCorporations[i].Signatories) > 0 && !lpa.TrustCorporations[i].Signatories[0].SignedAt.IsZero() {
				return []shared.FieldError{{Source: "/type", Detail: "trust corporation cannot opt out after signing"}}
			}

			lpa.TrustCorporations[i].Status = shared.AttorneyStatusRemoved
			return nil
		}
	}

	return []shared.FieldError{{Source: "/type", Detail: "trust corporation not found"}}
}

func validateTrustCorporationOptOut(update shared.Update) (TrustCorporationOptOut, []shared.FieldError) {
	if len(update.Changes) > 0 {
		return TrustCorporationOptOut{}, []shared.FieldError{{Source: "/changes", Detail: "expected empty"}}
	}

	author := update.Author.Details()

	if errs := validate.UUID("/author", author.UID); len(errs) > 0 {
		return TrustCorporationOptOut{}, errs
	}

	return TrustCorporationOptOut{trustCorporationUID: author.UID}, nil
}
