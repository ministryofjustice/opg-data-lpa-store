package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
	"time"
)

type Correction struct {
	DonorFirstNames     string
	DonorLastName       string
	DonorOtherNames     string
	DonorDob            shared.Date
	DonorAddress        shared.Address
	DonorEmail          string
	CertificateProvider CertificateProviderCorrection
	LPASignedAt         time.Time
}

type CertificateProviderCorrection struct {
	FirstNames string
	LastName   string
	Address    shared.Address
	Email      string
	Phone      string
	SignedAt   time.Time
}

func (cpr CertificateProviderCorrection) Apply(lpa *shared.Lpa) {
	lpa.CertificateProvider.FirstNames = cpr.FirstNames
	lpa.CertificateProvider.LastName = cpr.LastName
	lpa.CertificateProvider.Address = cpr.Address
	lpa.CertificateProvider.Email = cpr.Email
	lpa.CertificateProvider.Phone = cpr.Phone
	lpa.CertificateProvider.SignedAt = &cpr.SignedAt
}

func (c Correction) Apply(lpa *shared.Lpa) []shared.FieldError {
	if !c.LPASignedAt.IsZero() && lpa.Channel == shared.ChannelOnline {
		return []shared.FieldError{{Source: "/signedAt", Detail: "LPA Signed on date cannot be changed for online LPAs"}}
	}

	if lpa.Status == shared.LpaStatusRegistered {
		return []shared.FieldError{{Source: "/type", Detail: "Cannot make corrections to a Registered LPA"}}
	}

	lpa.Donor.FirstNames = c.DonorFirstNames
	lpa.Donor.LastName = c.DonorLastName
	lpa.Donor.OtherNamesKnownBy = c.DonorOtherNames
	lpa.Donor.DateOfBirth = c.DonorDob
	lpa.Donor.Address = c.DonorAddress
	lpa.Donor.Email = c.DonorEmail
	lpa.SignedAt = c.LPASignedAt

	return nil
}

func validateCorrection(changes []shared.Change, lpa *shared.Lpa) (Correction, []shared.FieldError) {
	var data Correction

	data.DonorFirstNames = lpa.LpaInit.Donor.FirstNames
	data.DonorLastName = lpa.LpaInit.Donor.LastName
	data.DonorOtherNames = lpa.LpaInit.Donor.OtherNamesKnownBy
	data.DonorDob = lpa.LpaInit.Donor.DateOfBirth
	data.DonorAddress = lpa.LpaInit.Donor.Address
	data.DonorEmail = lpa.LpaInit.Donor.Email
	data.LPASignedAt = lpa.LpaInit.SignedAt

	errors := parse.Changes(changes).
		Prefix("/donor/address", func(p *parse.Parser) []shared.FieldError {
			return p.
				Field("/line1", &data.DonorAddress.Line1, parse.Optional()).
				Field("/line2", &data.DonorAddress.Line2, parse.Optional()).
				Field("/line3", &data.DonorAddress.Line3, parse.Optional()).
				Field("/town", &data.DonorAddress.Town, parse.Optional()).
				Field("/postcode", &data.DonorAddress.Postcode, parse.Optional()).
				Field("/country", &data.DonorAddress.Country, parse.Validate(func() []shared.FieldError {
					return validate.Country("", data.DonorAddress.Country)
				}), parse.Optional()).
				Consumed()
		}, parse.Optional()).
		Field("/donor/firstNames", &data.DonorFirstNames, parse.Validate(func() []shared.FieldError {
			return validate.Required("", data.DonorFirstNames)
		}), parse.Optional()).
		Field("/donor/lastName", &data.DonorLastName, parse.Validate(func() []shared.FieldError {
			return validate.Required("", data.DonorLastName)
		}), parse.Optional()).
		Field("/donor/otherNamesKnownBy", &data.DonorOtherNames, parse.Optional()).
		Field("/donor/email", &data.DonorEmail, parse.Optional()).
		Field("/donor/dateOfBirth", &data.DonorDob, parse.Validate(func() []shared.FieldError {
			return validate.Date("", data.DonorDob)
		}), parse.Optional()).
		Field("/signedAt", &data.LPASignedAt, parse.Validate(func() []shared.FieldError {
			return validate.Time("", data.LPASignedAt)
		}), parse.Optional()).
		Consumed()

	return data, errors
}
func validateCertificateProvider(certificateProvider *CertificateProviderCorrection) func(p *parse.Parser) []shared.FieldError {
	return func(p *parse.Parser) []shared.FieldError {
		return p.
			Field("/firstNames", &certificateProvider.FirstNames, parse.Validate(func() []shared.FieldError {
				return validate.Required("", certificateProvider.FirstNames)
			}), parse.Optional()).
			Field("/lastName", &certificateProvider.LastName, parse.Validate(func() []shared.FieldError {
				return validate.Required("", certificateProvider.LastName)
			}), parse.Optional()).
			Prefix("/address", validateAddress(&certificateProvider.Address), parse.Optional()).
			Field("/email", &certificateProvider.Email, parse.Optional()).
			Field("/phone", &certificateProvider.Phone, parse.Optional()).
			Field("/signedAt", &certificateProvider.SignedAt, parse.Validate(func() []shared.FieldError {
				return validate.Time("", certificateProvider.SignedAt)
			}), parse.Optional()).
			Consumed()

	}
}

func validateAddress(address *shared.Address) func(p *parse.Parser) []shared.FieldError {
	return func(p *parse.Parser) []shared.FieldError {
		return p.
			Field("/line1", &address.Line1, parse.Optional()).
			Field("/line2", &address.Line2, parse.Optional()).
			Field("/line3", &address.Line3, parse.Optional()).
			Field("/town", &address.Town, parse.Optional()).
			Field("/postcode", &address.Postcode, parse.Optional()).
			Field("/country", &address.Country, parse.Validate(func() []shared.FieldError {
				return validate.Country("", address.Country)
			}), parse.Optional()).
			Consumed()
	}
}
