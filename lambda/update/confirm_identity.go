package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
)

type IdCheckComplete struct {
	Actor         idccActor
	IdentityCheck *shared.IdentityCheck
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

func validateConfirmIdentity(prefix string, actor idccActor, ic *shared.IdentityCheck, changes []shared.Change) (IdCheckComplete, []shared.FieldError) {
	var idcc IdCheckComplete

	errors := parse.Changes(changes).
		Prefix(prefix, func(p *parse.Parser) []shared.FieldError {
			idcc.Actor = actor

			if ic == nil {
				ic = &shared.IdentityCheck{}
			}
			idcc.IdentityCheck = ic

			return p.
				Field("/type", &ic.Type, parse.Validate(func() []shared.FieldError {
					return validate.IsValid("", ic.Type)
				})).
				Field("/checkedAt", &ic.CheckedAt, parse.Validate(func() []shared.FieldError {
					return validate.Time("", ic.CheckedAt)
				})).
				Consumed()
		}).
		Consumed()

	return idcc, errors
}

func validateDonorConfirmIdentity(changes []shared.Change, lpa *shared.Lpa) (IdCheckComplete, []shared.FieldError) {
	return validateConfirmIdentity("/donor/identityCheck", donor, lpa.Donor.IdentityCheck, changes)
}

func validateCertificateProviderConfirmIdentity(changes []shared.Change, lpa *shared.Lpa) (IdCheckComplete, []shared.FieldError) {
	return validateConfirmIdentity("/certificateProvider/identityCheck", certificateProvider, lpa.CertificateProvider.IdentityCheck, changes)
}
