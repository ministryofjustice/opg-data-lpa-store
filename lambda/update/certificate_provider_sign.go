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
	lpa.CertificateProvider.Channel = c.Channel

	return nil
}

func validateCertificateProviderSign(changes []shared.Change, lpa *shared.Lpa) (CertificateProviderSign, []shared.FieldError) {
	data := CertificateProviderSign{
		Address:                   lpa.LpaInit.CertificateProvider.Address,
		ContactLanguagePreference: lpa.LpaInit.CertificateProvider.ContactLanguagePreference,
		Email:                     lpa.LpaInit.CertificateProvider.Email,
		Channel:                   lpa.LpaInit.CertificateProvider.Channel,
	}

	if lpa.LpaInit.CertificateProvider.SignedAt != nil {
		data.SignedAt = *lpa.LpaInit.CertificateProvider.SignedAt
	}

	errors := parse.Changes(changes).
		Prefix("/certificateProvider/address", func(p *parse.Parser) []shared.FieldError {
			return p.
				Field("/line1", &data.Address.Line1, parse.MustMatchExisting()).
				Field("/line2", &data.Address.Line2, parse.Optional(), parse.MustMatchExisting()).
				Field("/line3", &data.Address.Line3, parse.Optional(), parse.MustMatchExisting()).
				Field("/town", &data.Address.Town, parse.MustMatchExisting()).
				Field("/postcode", &data.Address.Postcode, parse.Optional(), parse.MustMatchExisting()).
				Field("/country", &data.Address.Country, parse.Validate(func() []shared.FieldError {
					return validate.Country("", data.Address.Country)
				}), parse.MustMatchExisting()).
				Consumed()
		}, parse.Optional()).
		Field("/certificateProvider/signedAt", &data.SignedAt, parse.Validate(func() []shared.FieldError {
			return validate.Time("", data.SignedAt)
		}), parse.MustMatchExisting()).
		Field("/certificateProvider/contactLanguagePreference", &data.ContactLanguagePreference, parse.Validate(func() []shared.FieldError {
			return validate.IsValid("", data.ContactLanguagePreference)
		}), parse.MustMatchExisting()).
		Field("/certificateProvider/email", &data.Email, parse.Validate(func() []shared.FieldError {
			return validate.Required("", data.Email)
		}), parse.Optional(), parse.MustMatchExisting()).
		Field("/certificateProvider/channel", &data.Channel, parse.Validate(func() []shared.FieldError {
			return validate.IsValid("", data.Channel)
		}), parse.Optional(), parse.MustMatchExisting()).
		Consumed()

	return data, errors
}
