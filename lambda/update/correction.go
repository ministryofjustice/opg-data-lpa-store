package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
	"strconv"
	"time"
)

type Correction struct {
	Donor               DonorCorrection
	Attorney            AttorneyCorrection
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

type DonorCorrection struct {
	FirstNames        string
	LastName          string
	OtherNamesKnownBy string
	DateOfBirth       shared.Date
	Address           shared.Address
	Email             string
}

type AttorneyCorrection struct {
	Index      *int
	FirstNames string
	LastName   string
	Dob        shared.Date
	Address    shared.Address
	Email      string
	Mobile     string
	SignedAt   time.Time
}

const signedAt = "/signedAt"

func (c Correction) Apply(lpa *shared.Lpa) []shared.FieldError {
	if !c.LPASignedAt.IsZero() && !c.LPASignedAt.Equal(lpa.SignedAt) && lpa.Channel == shared.ChannelOnline {
		return []shared.FieldError{{Source: signedAt, Detail: "LPA Signed on date cannot be changed for online LPAs"}}
	}

	if c.Attorney.Index != nil && !c.Attorney.SignedAt.IsZero() && !c.Attorney.SignedAt.Equal(*lpa.Attorneys[*c.Attorney.Index].SignedAt) && lpa.Channel == shared.ChannelOnline {
		source := "/attorney/" + strconv.Itoa(*c.Attorney.Index) + signedAt
		return []shared.FieldError{{Source: source, Detail: "The attorney signed at date cannot be changed for online LPA"}}
	}

	if lpa.Status == shared.LpaStatusRegistered {
		return []shared.FieldError{{Source: "/type", Detail: "Cannot make corrections to a Registered LPA"}}
	}

	c.Donor.Apply(lpa)
	c.Attorney.Apply(lpa)
	lpa.SignedAt = c.LPASignedAt

	c.CertificateProvider.Apply(lpa)

	return nil
}

func (d DonorCorrection) Apply(lpa *shared.Lpa) {
	lpa.Donor.FirstNames = d.FirstNames
	lpa.Donor.LastName = d.LastName
	lpa.Donor.OtherNamesKnownBy = d.OtherNamesKnownBy
	lpa.Donor.DateOfBirth = d.DateOfBirth
	lpa.Donor.Address = d.Address
	lpa.Donor.Email = d.Email
}

func (a AttorneyCorrection) Apply(lpa *shared.Lpa) {
	if a.Index != nil {
		attorney := &lpa.Attorneys[*a.Index]
		attorney.FirstNames = a.FirstNames
		attorney.LastName = a.LastName
		attorney.DateOfBirth = a.Dob
		attorney.Address = a.Address
		attorney.Email = a.Email
		attorney.Mobile = a.Mobile
		attorney.SignedAt = &a.SignedAt
	}
}

func validateCorrection(changes []shared.Change, lpa *shared.Lpa) (Correction, []shared.FieldError) {
	var data Correction

	data.Donor.FirstNames = lpa.LpaInit.Donor.FirstNames
	data.Donor.LastName = lpa.LpaInit.Donor.LastName
	data.Donor.OtherNamesKnownBy = lpa.LpaInit.Donor.OtherNamesKnownBy
	data.Donor.DateOfBirth = lpa.LpaInit.Donor.DateOfBirth
	data.Donor.Address = lpa.LpaInit.Donor.Address
	data.Donor.Email = lpa.LpaInit.Donor.Email
	data.LPASignedAt = lpa.LpaInit.SignedAt

	data.CertificateProvider.FirstNames = lpa.LpaInit.CertificateProvider.FirstNames
	data.CertificateProvider.LastName = lpa.LpaInit.CertificateProvider.LastName
	data.CertificateProvider.Address = lpa.LpaInit.CertificateProvider.Address
	data.CertificateProvider.Email = lpa.LpaInit.CertificateProvider.Email
	data.CertificateProvider.Phone = lpa.LpaInit.CertificateProvider.Phone
	data.CertificateProvider.SignedAt = lpa.LpaInit.SignedAt

	errors := parse.Changes(changes).
		Prefix("/donor", validateDonor(&data.Donor), parse.Optional()).
		Field(signedAt, &data.LPASignedAt, parse.Validate(func() []shared.FieldError {
			return validate.Time("", data.LPASignedAt)
		}), parse.Optional()).
		Prefix("/attorneys", func(p *parse.Parser) []shared.FieldError {
			return p.
				Each(func(i int, p *parse.Parser) []shared.FieldError {
					if data.Attorney.Index != nil && *data.Attorney.Index != i {
						return p.OutOfRange()
					}

					data.Attorney.Index = &i
					data.Attorney.FirstNames = lpa.Attorneys[i].FirstNames
					data.Attorney.LastName = lpa.Attorneys[i].LastName
					data.Attorney.Dob = lpa.Attorneys[i].DateOfBirth
					data.Attorney.Address = lpa.Attorneys[i].Address
					data.Attorney.Email = lpa.Attorneys[i].Email
					data.Attorney.Mobile = lpa.Attorneys[i].Mobile

					if lpa.Attorneys[i].SignedAt != nil {
						data.Attorney.SignedAt = *lpa.Attorneys[i].SignedAt
					}

					return validateAttorney(&data.Attorney, p)
				}).
				Consumed()
		}, parse.Optional()).
		Prefix("/certificateProvider", validateCertificateProvider(&data.CertificateProvider), parse.Optional()).
		Consumed()

	return data, errors
}

func validateAttorney(attorney *AttorneyCorrection, p *parse.Parser) []shared.FieldError {
	return p.
		Field("/firstNames", &attorney.FirstNames, parse.Validate(func() []shared.FieldError {
			return validate.Required("", attorney.FirstNames)
		}), parse.Optional()).
		Field("/lastName", &attorney.LastName, parse.Validate(func() []shared.FieldError {
			return validate.Required("", attorney.LastName)
		}), parse.Optional()).
		Field("/dateOfBirth", &attorney.Dob, parse.Validate(func() []shared.FieldError {
			return validate.Date("", attorney.Dob)
		}), parse.Optional()).
		Field("/email", &attorney.Email, parse.Optional()).
		Field("/mobile", &attorney.Mobile, parse.Optional()).
		Prefix("/address", validateAddress(&attorney.Address), parse.Optional()).
		Field(signedAt, &attorney.SignedAt, parse.Validate(func() []shared.FieldError {
			return validate.Time("", attorney.SignedAt)
		}), parse.Optional()).
		Consumed()
}

func validateDonor(donor *DonorCorrection) func(p *parse.Parser) []shared.FieldError {
	return func(p *parse.Parser) []shared.FieldError {
		return p.
			Field("/firstNames", &donor.FirstNames, parse.Validate(func() []shared.FieldError {
				return validate.Required("", donor.FirstNames)
			}), parse.Optional()).
			Field("/lastName", &donor.LastName, parse.Validate(func() []shared.FieldError {
				return validate.Required("", donor.LastName)
			}), parse.Optional()).
			Field("/otherNamesKnownBy", &donor.OtherNamesKnownBy, parse.Optional()).
			Field("/dateOfBirth", &donor.DateOfBirth, parse.Validate(func() []shared.FieldError {
				return validate.Date("", donor.DateOfBirth)
			}), parse.Optional()).
			Prefix("/address", validateAddress(&donor.Address), parse.Optional()).
			Field("/email", &donor.Email, parse.Optional()).
			Consumed()
	}
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
