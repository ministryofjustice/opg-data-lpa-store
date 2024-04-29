package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
)

type TrustCorporationSign struct {
	Index                     *int
	Mobile                    string
	Signatories               [2]shared.Signatory
	ContactLanguagePreference shared.Lang
}

func (a TrustCorporationSign) Apply(lpa *shared.Lpa) []shared.FieldError {
	if signatories := lpa.TrustCorporations[*a.Index].Signatories; len(signatories) > 0 && !signatories[0].SignedAt.IsZero() {
		return []shared.FieldError{{Source: "/type", Detail: "trust corporation cannot sign again"}}
	}

	lpa.TrustCorporations[*a.Index].Mobile = a.Mobile
	if a.Signatories[1].IsZero() {
		lpa.TrustCorporations[*a.Index].Signatories = a.Signatories[:1]
	} else {
		lpa.TrustCorporations[*a.Index].Signatories = a.Signatories[:]
	}
	lpa.TrustCorporations[*a.Index].ContactLanguagePreference = a.ContactLanguagePreference

	return nil
}

func validateTrustCorporationSign(changes []shared.Change, lpa *shared.Lpa) (TrustCorporationSign, []shared.FieldError) {
	var data TrustCorporationSign

	errors := parse.Changes(changes).
		Prefix("/trustCorporations", func(prefix *parse.Parser) []shared.FieldError {
			return prefix.
				Each(func(i int, each *parse.Parser) []shared.FieldError {
					if data.Index != nil && *data.Index != i {
						return each.OutOfRange()
					}

					data.Index = &i
					return each.
						Field("/mobile", &data.Mobile).
						Field("/contactLanguagePreference", &data.ContactLanguagePreference, parse.Validate(func() []shared.FieldError {
							return validate.IsValid("", data.ContactLanguagePreference)
						})).
						Prefix("/signatories", func(prefix *parse.Parser) []shared.FieldError {
							return prefix.
								Each(func(i int, each *parse.Parser) []shared.FieldError {
									if i > 1 {
										return each.OutOfRange()
									}

									return each.
										Field("/firstNames", &data.Signatories[i].FirstNames).
										Field("/lastName", &data.Signatories[i].LastName).
										Field("/professionalTitle", &data.Signatories[i].ProfessionalTitle).
										Field("/signedAt", &data.Signatories[i].SignedAt).
										Consumed()
								}, 0).
								Consumed()
						}).
						Consumed()
				}).
				Consumed()
		}).
		Consumed()

	return data, errors
}
