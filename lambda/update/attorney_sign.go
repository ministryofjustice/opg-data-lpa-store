package main

import (
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
)

type AttorneySign struct {
	Index                     *int
	Mobile                    string
	SignedAt                  time.Time
	ContactLanguagePreference shared.Lang
}

func (a AttorneySign) Apply(lpa *shared.Lpa) []shared.FieldError {
	if !lpa.Attorneys[*a.Index].SignedAt.IsZero() {
		return []shared.FieldError{{Source: "/type", Detail: "attorney cannot sign again"}}
	}

	lpa.Attorneys[*a.Index].Mobile = a.Mobile
	lpa.Attorneys[*a.Index].SignedAt = a.SignedAt
	lpa.Attorneys[*a.Index].ContactLanguagePreference = a.ContactLanguagePreference

	return nil
}

func validateAttorneySign(changes []shared.Change) (AttorneySign, []shared.FieldError) {
	var data AttorneySign

	errors := parse.Changes(changes).
		Prefix("/attorneys", func(p *parse.Parser) []shared.FieldError {
			return p.
				Each(func(i int, p *parse.Parser) []shared.FieldError {
					if data.Index != nil && *data.Index != i {
						return p.OutOfRange()
					}

					data.Index = &i
					return p.
						Field("/mobile", &data.Mobile).
						ValidatedField("/signedAt", &data.SignedAt, func() []shared.FieldError {
							return validate.Time("", data.SignedAt)
						}).
						ValidatedField("/contactLanguagePreference", &data.ContactLanguagePreference, func() []shared.FieldError {
							return validate.IsValid("", data.ContactLanguagePreference)
						}).
						Consumed()
				}).
				Consumed()
		}).
		Consumed()

	return data, errors
}
