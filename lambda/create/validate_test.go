package main

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/lambda/shared"
	"github.com/stretchr/testify/assert"
)

var validAddress = shared.Address{
	Line1:   "123 Main St",
	Town:    "Homeland",
	Country: "GB",
}

func newDate(date string, isMalformed bool) shared.Date {
	t, _ := time.Parse("2006-01-02", date)

	return shared.Date{
		Time:        t,
		IsMalformed: isMalformed,
	}
}

func TestValidateAddressEmpty(t *testing.T) {
	address := shared.Address{}
	errors := validateAddress(address, "/test", []shared.FieldError{})

	assert.Contains(t, errors, shared.FieldError{Source: "/test/line1", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/town", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/country", Detail: "field is required"})
}

func TestValidateAddressValid(t *testing.T) {
	errors := validateAddress(validAddress, "/test", []shared.FieldError{})

	assert.Empty(t, errors)
}

func TestValidateAddressInvalidCountry(t *testing.T) {
	invalidAddress := shared.Address{
		Line1:   "123 Main St",
		Town:    "Homeland",
		Country: "United Kingdom",
	}
	errors := validateAddress(invalidAddress, "/test", []shared.FieldError{})

	assert.Contains(t, errors, shared.FieldError{Source: "/test/country", Detail: "must be a valid ISO-3166-1 country code"})
}

func TestValidateAttorneyEmpty(t *testing.T) {
	attorney := shared.Attorney{}
	errors := validateAttorney(attorney, "/test", []shared.FieldError{})

	assert.Contains(t, errors, shared.FieldError{Source: "/test/firstNames", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/surname", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/status", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/dateOfBirth", Detail: "field is required"})
}

func TestValidateAttorneyValid(t *testing.T) {
	attorney := shared.Attorney{
		Person: shared.Person{
			FirstNames:  "Lesia",
			Surname:     "Lathim",
			Address:     validAddress,
			DateOfBirth: newDate("1928-01-18", false),
		},
		Status: shared.AttorneyStatusActive,
	}
	errors := validateAttorney(attorney, "/test", []shared.FieldError{})

	assert.Empty(t, errors)
}

func TestValidateAttorneyMalformedDateOfBirth(t *testing.T) {
	attorney := shared.Attorney{
		Person: shared.Person{
			FirstNames:  "Lesia",
			Surname:     "Lathim",
			Address:     validAddress,
			DateOfBirth: newDate("bad date", true),
		},
		Status: shared.AttorneyStatusActive,
	}
	errors := validateAttorney(attorney, "/test", []shared.FieldError{})

	assert.Contains(t, errors, shared.FieldError{Source: "/test/dateOfBirth", Detail: "invalid format"})
}

func TestValidateAttorneyInvalidStatus(t *testing.T) {
	attorney := shared.Attorney{
		Person: shared.Person{
			FirstNames:  "Lesia",
			Surname:     "Lathim",
			Address:     validAddress,
			DateOfBirth: newDate("1928-01-18", false),
		},
		Status: "bad status",
	}
	errors := validateAttorney(attorney, "/test", []shared.FieldError{})

	assert.Contains(t, errors, shared.FieldError{Source: "/test/status", Detail: "invalid value"})
}

func TestValidateLpaEmpty(t *testing.T) {
	lpa := shared.LpaInit{}
	errors := Validate(lpa)

	assert.Contains(t, errors, shared.FieldError{Source: "/donor/firstNames", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/donor/surname", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/donor/dateOfBirth", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/attorneys", Detail: "at least one attorney is required"})
}

func TestValidateLpaValid(t *testing.T) {
	lpa := shared.LpaInit{
		Donor: shared.Donor{
			Person: shared.Person{
				FirstNames:  "Otto",
				Surname:     "Boudreau",
				DateOfBirth: newDate("1956-08-08", false),
				Address:     validAddress,
			},
		},
		Attorneys: []shared.Attorney{
			{
				Person: shared.Person{
					FirstNames:  "Sharonda",
					Surname:     "Graciani",
					DateOfBirth: newDate("1977-10-30", false),
					Address:     validAddress,
				},
				Status: shared.AttorneyStatusActive,
			},
		},
	}
	errors := Validate(lpa)

	assert.Empty(t, errors)
}
