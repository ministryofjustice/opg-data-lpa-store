package main

import (
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
)

type CertificateProviderSign struct {
	Address                   shared.Address
	SignedAt                  time.Time
	ContactLanguagePreference shared.Lang
}

func (c CertificateProviderSign) Apply(lpa *shared.Lpa) []shared.FieldError {
	if !lpa.CertificateProvider.SignedAt.IsZero() {
		return []shared.FieldError{{Source: "/type", Detail: "certificate provider cannot sign again"}}
	}

	lpa.CertificateProvider.Address = c.Address
	lpa.CertificateProvider.SignedAt = c.SignedAt
	lpa.CertificateProvider.ContactLanguagePreference = c.ContactLanguagePreference

	return nil
}

func validateCertificateProviderSign(changes []shared.Change) (CertificateProviderSign, []shared.FieldError) {
	var data CertificateProviderSign

	errors := parse.Changes(changes).
		OptionalPrefix("/certificateProvider/address", func(p *parse.Parser) []shared.FieldError {
			return p.
				Field("/line1", &data.Address.Line1).
				OptionalField("/line2", &data.Address.Line2).
				OptionalField("/line3", &data.Address.Line3).
				Field("/town", &data.Address.Town).
				OptionalField("/postcode", &data.Address.Postcode).
				ValidatedField("/country", &data.Address.Country, func() []shared.FieldError {
					return validate.Country("", data.Address.Country)
				}).
				Consumed()
		}).
		ValidatedField("/certificateProvider/signedAt", &data.SignedAt, func() []shared.FieldError {
			return validate.Time("", data.SignedAt)
		}).
		ValidatedField("/certificateProvider/contactLanguagePreference", &data.ContactLanguagePreference, func() []shared.FieldError {
			return validate.IsValid("", data.ContactLanguagePreference)
		}).
		Consumed()

	return data, errors
}
