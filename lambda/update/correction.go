package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
	"time"
)

type Correction struct {
	DonorFirstNames string
	DonorLastName   string
	DonorOtherNames string
	DonorDob        shared.Date
	DonorAddress    shared.Address
	DonorEmail      string
	LPASignedAt     time.Time
}

func (c Correction) Apply(lpa *shared.Lpa) []shared.FieldError {

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
