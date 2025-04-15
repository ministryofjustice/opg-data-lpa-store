package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
)

type PaperAttorneyAccessOnline struct {
	Index *int
	Email string
}

func (a PaperAttorneyAccessOnline) Apply(lpa *shared.Lpa) []shared.FieldError {
	if lpa.Attorneys[*a.Index].Channel != shared.ChannelPaper {
		return []shared.FieldError{{Source: "/channel", Detail: "lpa channel is not paper"}}
	}

	lpa.Attorneys[*a.Index].Email = a.Email

	return nil
}

func validatePaperAttorneyAccessOnline(changes []shared.Change, lpa *shared.Lpa) (Applyable, []shared.FieldError) {
	var data PaperAttorneyAccessOnline

	errors := parse.Changes(changes).
		Prefix("/attorneys", func(p *parse.Parser) []shared.FieldError {
			return p.
				EachKey(func(key string, p *parse.Parser) []shared.FieldError {
					attorneyIdx, ok := lpa.FindAttorneyIndex(key)

					if !ok || (data.Index != nil && *data.Index != attorneyIdx) {
						return p.OutOfRange()
					}

					data.Index = &attorneyIdx

					return p.
						Field("/email", &data.Email, parse.Validate(validate.NotEmpty())).
						Consumed()
				}).
				Consumed()
		}).
		Consumed()

	return data, errors
}
