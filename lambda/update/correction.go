package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
	"strconv"
	"time"
)

type Correction struct {
	LPASignedAt        time.Time
	Index              *int
	AttorneyFirstNames string
	AttorneyLastName   string
	AttorneyDob        shared.Date
	AttorneyAddress    shared.Address
	AttorneyEmail      string
	AttorneyMobile     string
	AttorneySignedAt   time.Time
	Donor              DonorCorrection
}

type DonorCorrection struct {
	FirstNames        string
	LastName          string
	OtherNamesKnownBy string
	DateOfBirth       shared.Date
	Address           shared.Address
	Email             string
}

const signedAt = "/signedAt"

func (c Correction) Apply(lpa *shared.Lpa) []shared.FieldError {
	if !c.LPASignedAt.IsZero() && lpa.Channel == shared.ChannelOnline {
		return []shared.FieldError{{Source: signedAt, Detail: "LPA Signed on date cannot be changed for online LPAs"}}
	}

	if c.Index != nil && lpa.Attorneys[*c.Index].SignedAt != nil && !lpa.Attorneys[*c.Index].SignedAt.IsZero() && lpa.Channel == shared.ChannelOnline {
		source := "/attorney/" + strconv.Itoa(*c.Index) + signedAt
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

	if c.Index != nil {
		lpa.Attorneys[*c.Index].FirstNames = c.AttorneyFirstNames
		lpa.Attorneys[*c.Index].LastName = c.AttorneyLastName
		lpa.Attorneys[*c.Index].DateOfBirth = c.AttorneyDob
		lpa.Attorneys[*c.Index].Address = c.AttorneyAddress
		lpa.Attorneys[*c.Index].Email = c.AttorneyEmail
		lpa.Attorneys[*c.Index].Mobile = c.AttorneyMobile
		lpa.Attorneys[*c.Index].SignedAt = &c.AttorneySignedAt
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
					if data.Index != nil && *data.Index != i {
						return p.OutOfRange()
					}

					data.Index = &i
					data.AttorneyFirstNames = lpa.Attorneys[i].FirstNames
					data.AttorneyLastName = lpa.Attorneys[i].LastName
					data.AttorneyDob = lpa.Attorneys[i].DateOfBirth
					data.AttorneyAddress = lpa.Attorneys[i].Address
					data.AttorneyEmail = lpa.Attorneys[i].Email
					data.AttorneyMobile = lpa.Attorneys[i].Mobile

					if lpa.Attorneys[i].SignedAt != nil {
						data.AttorneySignedAt = *lpa.Attorneys[i].SignedAt
					}

					return p.
						Field("/firstNames", &data.AttorneyFirstNames, parse.Validate(func() []shared.FieldError {
							return validate.Required("", data.AttorneyFirstNames)
						}), parse.Optional()).
						Field("/lastName", &data.AttorneyLastName, parse.Validate(func() []shared.FieldError {
							return validate.Required("", data.AttorneyLastName)
						}), parse.Optional()).
						Field("/dateOfBirth", &data.AttorneyDob, parse.Validate(func() []shared.FieldError {
							return validate.Date("", data.AttorneyDob)
						}), parse.Optional()).
						Field("/email", &data.AttorneyEmail, parse.Optional()).
						Field("/mobile", &data.AttorneyMobile, parse.Optional()).
						Prefix("/address", validateAddress(&data.AttorneyAddress), parse.Optional()).
						Field(signedAt, &data.AttorneySignedAt, parse.Validate(func() []shared.FieldError {
							return validate.Time("", data.AttorneySignedAt)
						}), parse.Optional()).
						Consumed()
				}).
				Consumed()
		}, parse.Optional()).
		Consumed()

	return data, errors
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
