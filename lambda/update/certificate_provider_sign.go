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
	Email                     string
	Channel                   shared.Channel
}

func (c CertificateProviderSign) Apply(lpa *shared.Lpa) []shared.FieldError {
	if lpa.CertificateProvider.SignedAt != nil && !lpa.CertificateProvider.SignedAt.IsZero() {
		return []shared.FieldError{{Source: "/type", Detail: "certificate provider cannot sign again"}}
	}

	lpa.CertificateProvider.Address = c.Address
	lpa.CertificateProvider.SignedAt = &c.SignedAt
	lpa.CertificateProvider.ContactLanguagePreference = c.ContactLanguagePreference
	lpa.CertificateProvider.Email = c.Email
	// to account for paper to online moves
	lpa.CertificateProvider.Channel = c.Channel

	return nil
}

func validateCertificateProviderSign(changes []shared.Change) (CertificateProviderSign, []shared.FieldError) {
	var data CertificateProviderSign

	errors := parse.Changes(changes).
		Prefix("/certificateProvider/address", func(p *parse.Parser) []shared.FieldError {
			return p.
				Field("/line1", &data.Address.Line1).
				Field("/line2", &data.Address.Line2, parse.Optional()).
				Field("/line3", &data.Address.Line3, parse.Optional()).
				Field("/town", &data.Address.Town).
				Field("/postcode", &data.Address.Postcode, parse.Optional()).
				Field("/country", &data.Address.Country, parse.Validate(func() []shared.FieldError {
					return validate.Country("", data.Address.Country)
				})).
				Consumed()
		}, parse.Optional()).
		Field("/certificateProvider/signedAt", &data.SignedAt, parse.Validate(func() []shared.FieldError {
			return validate.Time("", data.SignedAt)
		})).
		Field("/certificateProvider/contactLanguagePreference", &data.ContactLanguagePreference, parse.Validate(func() []shared.FieldError {
			return validate.IsValid("", data.ContactLanguagePreference)
		})).
		Field("/certificateProvider/email", &data.Email, parse.Validate(func() []shared.FieldError {
			return validate.Required("", data.Email)
		}), parse.Optional(), parse.UpdateExisting()).
		Field("/certificateProvider/channel", &data.Channel, parse.Validate(func() []shared.FieldError {
			return validate.IsValid("", data.Channel)
		}), parse.Optional(), parse.UpdateExisting()).
		Consumed()

	return data, errors
}
