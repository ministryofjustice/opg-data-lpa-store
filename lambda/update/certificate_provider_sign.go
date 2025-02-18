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
				Field("/line1", &data.Address.Line1).
				Field("/line2", &data.Address.Line2, parse.Optional()).
				Field("/line3", &data.Address.Line3, parse.Optional()).
				Field("/town", &data.Address.Town).
				Field("/postcode", &data.Address.Postcode, parse.Optional()).
				Field("/country", &data.Address.Country, parse.Validate(validate.Country())).
				Consumed()
		}, parse.Optional()).
		Field("/certificateProvider/signedAt", &data.SignedAt, parse.Validate(validate.NotEmpty())).
		Field("/certificateProvider/contactLanguagePreference", &data.ContactLanguagePreference, parse.Validate(validate.Valid())).
		Field("/certificateProvider/email", &data.Email, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Field("/certificateProvider/channel", &data.Channel, parse.Validate(validate.Valid()), parse.Optional()).
		Consumed()

	return data, errors
}
