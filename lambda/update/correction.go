package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
	"strconv"
	"time"
)

type Correction struct {
	Donor    DonorCorrection
	Attorney AttorneyCorrection
}

type DonorCorrection struct {
	FirstNames  string
	LastName    string
	OtherNames  string
	Dob         shared.Date
	Address     shared.Address
	Email       string
	LpaSignedAt time.Time
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
	var errors []shared.FieldError

	donorErrors := c.Donor.Apply(lpa)
	errors = append(errors, donorErrors...)

	if c.Attorney.Index != nil {
		attorneyErrors := c.Attorney.Apply(lpa)
		errors = append(errors, attorneyErrors...)
	}

	return errors
}

func (c DonorCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	if !c.LpaSignedAt.IsZero() && lpa.Channel == shared.ChannelOnline {
		return []shared.FieldError{{Source: signedAt, Detail: "LPA Signed on date cannot be changed for online LPAs"}}
	}

	if lpa.Status == shared.LpaStatusRegistered {
		return []shared.FieldError{{Source: "/type", Detail: "Cannot make corrections to a Registered LPA"}}
	}

	lpa.Donor.FirstNames = c.FirstNames
	lpa.Donor.LastName = c.LastName
	lpa.Donor.OtherNamesKnownBy = c.OtherNames
	lpa.Donor.DateOfBirth = c.Dob
	lpa.Donor.Address = c.Address
	lpa.Donor.Email = c.Email
	lpa.SignedAt = c.LpaSignedAt

	return nil
}

func (c AttorneyCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {

	if c.Index != nil && lpa.Attorneys[*c.Index].SignedAt != nil && !lpa.Attorneys[*c.Index].SignedAt.IsZero() && lpa.Channel == shared.ChannelOnline {
		source := "/attorney/" + strconv.Itoa(*c.Index) + signedAt
		return []shared.FieldError{{Source: source, Detail: "The attorney signed at date cannot be changed for online LPA"}}
	}

	if lpa.Status == shared.LpaStatusRegistered {
		return []shared.FieldError{{Source: "/type", Detail: "Cannot make corrections to a Registered LPA"}}
	}

	if c.Index != nil {
		lpa.Attorneys[*c.Index].FirstNames = c.FirstNames
		lpa.Attorneys[*c.Index].LastName = c.LastName
		lpa.Attorneys[*c.Index].DateOfBirth = c.Dob
		lpa.Attorneys[*c.Index].Address = c.Address
		lpa.Attorneys[*c.Index].Email = c.Email
		lpa.Attorneys[*c.Index].Mobile = c.Mobile
		lpa.Attorneys[*c.Index].SignedAt = &c.SignedAt
	}

	return nil
}

func validateCorrection(changes []shared.Change, lpa *shared.Lpa) (Correction, []shared.FieldError) {
	var data Correction

	data.Donor.FirstNames = lpa.LpaInit.Donor.FirstNames
	data.Donor.LastName = lpa.LpaInit.Donor.LastName
	data.Donor.OtherNames = lpa.LpaInit.Donor.OtherNamesKnownBy
	data.Donor.Dob = lpa.LpaInit.Donor.DateOfBirth
	data.Donor.Address = lpa.LpaInit.Donor.Address
	data.Donor.Email = lpa.LpaInit.Donor.Email
	data.Donor.LpaSignedAt = lpa.LpaInit.SignedAt

	errors := parse.Changes(changes).
		Prefix("/donor/address", validateAddress(&data.Donor.Address), parse.Optional()).
		Field("/donor/firstNames", &data.Donor.FirstNames, parse.Validate(func() []shared.FieldError {
			return validate.Required("", data.Donor.FirstNames)
		}), parse.Optional()).
		Field("/donor/lastName", &data.Donor.LastName, parse.Validate(func() []shared.FieldError {
			return validate.Required("", data.Donor.LastName)
		}), parse.Optional()).
		Field("/donor/otherNamesKnownBy", &data.Donor.OtherNames, parse.Optional()).
		Field("/donor/email", &data.Donor.Email, parse.Optional()).
		Field("/donor/dateOfBirth", &data.Donor.Dob, parse.Validate(func() []shared.FieldError {
			return validate.Date("", data.Donor.Dob)
		}), parse.Optional()).
		Field(signedAt, &data.Donor.LpaSignedAt, parse.Validate(func() []shared.FieldError {
			return validate.Time("", data.Donor.LpaSignedAt)
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

					return p.
						Field("/firstNames", &data.Attorney.FirstNames, parse.Validate(func() []shared.FieldError {
							return validate.Required("", data.Attorney.FirstNames)
						}), parse.Optional()).
						Field("/lastName", &data.Attorney.LastName, parse.Validate(func() []shared.FieldError {
							return validate.Required("", data.Attorney.LastName)
						}), parse.Optional()).
						Field("/dateOfBirth", &data.Attorney.Dob, parse.Validate(func() []shared.FieldError {
							return validate.Date("", data.Attorney.Dob)
						}), parse.Optional()).
						Field("/email", &data.Attorney.Email, parse.Optional()).
						Field("/mobile", &data.Attorney.Mobile, parse.Optional()).
						Prefix("/address", validateAddress(&data.Attorney.Address), parse.Optional()).
						Field(signedAt, &data.Attorney.SignedAt, parse.Validate(func() []shared.FieldError {
							return validate.Time("", data.Attorney.SignedAt)
						}), parse.Optional()).
						Consumed()
				}).
				Consumed()
		}, parse.Optional()).
		Consumed()

	return data, errors
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
