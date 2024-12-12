package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
	"strconv"
	"time"
)

type Correction struct {
	DonorFirstNames    string
	DonorLastName      string
	DonorOtherNames    string
	DonorDob           shared.Date
	DonorAddress       shared.Address
	DonorEmail         string
	LPASignedAt        time.Time
	Index              *int
	AttorneyFirstNames string
	AttorneyLastName   string
	AttorneyDob        shared.Date
	AttorneyAddress    shared.Address
	AttorneyEmail      string
	AttorneyMobile     string
	AttorneySignedAt   time.Time
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

	lpa.Donor.FirstNames = c.DonorFirstNames
	lpa.Donor.LastName = c.DonorLastName
	lpa.Donor.OtherNamesKnownBy = c.DonorOtherNames
	lpa.Donor.DateOfBirth = c.DonorDob
	lpa.Donor.Address = c.DonorAddress
	lpa.Donor.Email = c.DonorEmail
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
						Prefix("/address", func(p *parse.Parser) []shared.FieldError {
							return p.
								Field("/line1", &data.AttorneyAddress.Line1, parse.Optional()).
								Field("/line2", &data.AttorneyAddress.Line2, parse.Optional()).
								Field("/line3", &data.AttorneyAddress.Line3, parse.Optional()).
								Field("/town", &data.AttorneyAddress.Town, parse.Optional()).
								Field("/postcode", &data.AttorneyAddress.Postcode, parse.Optional()).
								Field("/country", &data.AttorneyAddress.Country, parse.Validate(func() []shared.FieldError {
									return validate.Country("", data.AttorneyAddress.Country)
								}), parse.Optional()).
								Consumed()
						}, parse.Optional()).
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
