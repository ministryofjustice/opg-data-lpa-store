package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
	"strconv"
	"time"
)

type Correction struct {
	Donor       DonorCorrection
	Attorney    AttorneyCorrection
	LPASignedAt time.Time
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
	if !c.LPASignedAt.IsZero() && lpa.Channel == shared.ChannelOnline {
		return []shared.FieldError{{Source: signedAt, Detail: "LPA Signed on date cannot be changed for online LPAs"}}
	}

	if c.Attorney.Index != nil && lpa.Attorneys[*c.Attorney.Index].SignedAt != nil && !lpa.Attorneys[*c.Attorney.Index].SignedAt.IsZero() && lpa.Channel == shared.ChannelOnline {
		source := "/attorney/" + strconv.Itoa(*c.Attorney.Index) + signedAt
		return []shared.FieldError{{Source: source, Detail: "The attorney signed at date cannot be changed for online LPA"}}
	}
	if lpa.Status == shared.LpaStatusRegistered {
		return []shared.FieldError{{Source: "/type", Detail: "Cannot make corrections to a Registered LPA"}}
	}

	lpa.Donor.FirstNames = c.Donor.FirstNames
	lpa.Donor.LastName = c.Donor.LastName
	lpa.Donor.OtherNamesKnownBy = c.Donor.OtherNamesKnownBy
	lpa.Donor.DateOfBirth = c.Donor.DateOfBirth
	lpa.Donor.Address = c.Donor.Address
	lpa.Donor.Email = c.Donor.Email
	lpa.SignedAt = c.LPASignedAt

	if c.Attorney.Index != nil {
		lpa.Attorneys[*c.Attorney.Index].FirstNames = c.Attorney.FirstNames
		lpa.Attorneys[*c.Attorney.Index].LastName = c.Attorney.LastName
		lpa.Attorneys[*c.Attorney.Index].DateOfBirth = c.Attorney.Dob
		lpa.Attorneys[*c.Attorney.Index].Address = c.Attorney.Address
		lpa.Attorneys[*c.Attorney.Index].Email = c.Attorney.Email
		lpa.Attorneys[*c.Attorney.Index].Mobile = c.Attorney.Mobile
		lpa.Attorneys[*c.Attorney.Index].SignedAt = &c.Attorney.SignedAt
	}

	return nil
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
