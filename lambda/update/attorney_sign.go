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
	Channel                   shared.Channel
	Email                     string
}

func (a AttorneySign) Apply(lpa *shared.Lpa) []shared.FieldError {
	if lpa.Attorneys[*a.Index].SignedAt != nil && !lpa.Attorneys[*a.Index].SignedAt.IsZero() {
		return []shared.FieldError{{Source: "/type", Detail: "attorney cannot sign again"}}
	}

	lpa.Attorneys[*a.Index].Mobile = a.Mobile
	lpa.Attorneys[*a.Index].SignedAt = &a.SignedAt
	lpa.Attorneys[*a.Index].ContactLanguagePreference = a.ContactLanguagePreference
	lpa.Attorneys[*a.Index].Channel = a.Channel
	lpa.Attorneys[*a.Index].Email = a.Email

	return nil
}

func validateAttorneySign(changes []shared.Change, lpa *shared.Lpa) (AttorneySign, []shared.FieldError) {
	var data AttorneySign

	errors := parse.Changes(changes).
		Prefix("/attorneys", func(p *parse.Parser) []shared.FieldError {
			return p.
				Each(func(i int, p *parse.Parser) []shared.FieldError {
					if data.Index != nil && *data.Index != i {
						return p.OutOfRange()
					}

					data.Index = &i
					data.Mobile = lpa.Attorneys[i].Mobile
					data.ContactLanguagePreference = lpa.Attorneys[i].ContactLanguagePreference
					data.Channel = lpa.Attorneys[i].Channel
					data.Email = lpa.Attorneys[i].Email

					if lpa.Attorneys[i].SignedAt != nil {
						data.SignedAt = *lpa.Attorneys[i].SignedAt
					}

					return p.
						Field("/mobile", &data.Mobile).
						Field("/signedAt", &data.SignedAt, parse.Validate(validate.NotEmpty())).
						Field("/contactLanguagePreference", &data.ContactLanguagePreference, parse.Validate(validate.Valid())).
						Field("/channel", &data.Channel, parse.Validate(validate.Valid()), parse.Optional()).
						Field("/email", &data.Email, parse.Optional()).
						Consumed()
				}).
				Consumed()
		}).
		Consumed()

	return data, errors
}
