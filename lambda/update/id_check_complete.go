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

func parseIdCheckCompleteChanges(prefix string, changes []shared.Change, existing IdCheckComplete) (IdCheckComplete, []shared.FieldError) {
	errors := parse.Changes(changes).
		Prefix(
			prefix,
			func(p *parse.Parser) []shared.FieldError {
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
			},
		).Errors()

	return existing, errors
}

func validateIdCheckComplete(changes []shared.Change, lpa *shared.Lpa) (IdCheckComplete, []shared.FieldError) {
	var existing IdCheckComplete

	// identify whether we are parsing donor or certificate provider identity check
	prefix := "/donor/identityCheck"
	errors := parse.Changes(changes).Prefix(prefix, func(p *parse.Parser) []shared.FieldError {
		// ignore validation on fields (for now)
		return []shared.FieldError{}
	}).Errors()

	// if we have errors at this point, we assume we are parsing a certificateProvider identity check
	if len(errors) > 0 {
		prefix = "/certificateProvider/identityCheck"

		// populate existing from certificate provider
		existing = IdCheckComplete{IdentityCheck: lpa.CertificateProvider.IdentityCheck, Actor: certificateProvider}
	} else {
		// populate existing from donor
		existing = IdCheckComplete{IdentityCheck: lpa.Donor.IdentityCheck, Actor: donor}
	}

	return parseIdCheckCompleteChanges(prefix, changes, existing)
}
