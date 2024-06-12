package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
)

type IdCheckComplete struct {
	Actor idccActor
	shared.IdentityCheck
}

type idccActor string

var (
	donor               = idccActor("Donor")
	certificateProvider = idccActor("CertificateProvider")
)

func (idcc IdCheckComplete) Apply(lpa *shared.Lpa) []shared.FieldError {
	if idcc.Actor == donor {
		lpa.Donor.IdentityCheck = idcc.IdentityCheck
	} else {
		lpa.CertificateProvider.IdentityCheck = idcc.IdentityCheck
	}
	return nil
}

func validateIdCheckComplete(changes []shared.Change, lpa *shared.Lpa) (IdCheckComplete, []shared.FieldError) {
	var existing IdCheckComplete

	identityCheckParser := func(actor idccActor) func(p *parse.Parser) []shared.FieldError {
		return func(p *parse.Parser) []shared.FieldError {
			if existing.Actor != "" {
				return []shared.FieldError{{Source: "/", Detail: "id check for multiple actors is not allowed"}}
			}

			switch actor {
			case donor:
				existing.IdentityCheck = lpa.Donor.IdentityCheck
			case certificateProvider:
				existing.IdentityCheck = lpa.CertificateProvider.IdentityCheck
			}

			existing.Actor = actor

			return p.
				Field("/type", &existing.Type, parse.Validate(func() []shared.FieldError {
					return validate.IsValid("", existing.Type)
				}), parse.MustMatchExisting()).
				Field("/date", &existing.CheckedAt, parse.Validate(func() []shared.FieldError {
					return validate.Time("", existing.CheckedAt)
				}), parse.MustMatchExisting()).
				Field("/reference", &existing.Reference, parse.Validate(func() []shared.FieldError {
					return validate.Required("", existing.Reference)
				}), parse.MustMatchExisting()).
				Consumed()
		}
	}

	errors := parse.Changes(changes).
		Prefix("/donor/identityCheck", identityCheckParser(donor), parse.Optional()).
		Prefix("/certificateProvider/identityCheck", identityCheckParser(certificateProvider), parse.Optional()).
		Errors()

	if existing.Actor == "" {
		return existing, append(errors, shared.FieldError{Source: "/", Detail: "id check for unknown actor type"})
	}

	return existing, errors
}
