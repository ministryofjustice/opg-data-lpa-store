package main

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
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

func TestCountAttorneys(t *testing.T) {
	actives, replacements := countAttorneys([]shared.Attorney{})
	assert.Equal(t, 0, actives)
	assert.Equal(t, 0, replacements)

	actives, replacements = countAttorneys([]shared.Attorney{
		{Status: shared.AttorneyStatusReplacement},
		{Status: shared.AttorneyStatusActive},
		{Status: shared.AttorneyStatusReplacement},
	})
	assert.Equal(t, 1, actives)
	assert.Equal(t, 2, replacements)
}

func TestFlatten(t *testing.T) {
	errA := shared.FieldError{Source: "a", Detail: "a"}
	errB := shared.FieldError{Source: "b", Detail: "b"}
	errC := shared.FieldError{Source: "c", Detail: "c"}

	assert.Nil(t, flatten())
	assert.Nil(t, flatten([]shared.FieldError{}, []shared.FieldError{}))
	assert.Equal(t, []shared.FieldError{errA, errB, errC}, flatten([]shared.FieldError{errA, errB}, []shared.FieldError{errC}))
	assert.Equal(t, []shared.FieldError{errA, errB, errC}, flatten([]shared.FieldError{errA}, []shared.FieldError{errB, errC}))
	assert.Equal(t, []shared.FieldError{errA, errB, errC}, flatten([]shared.FieldError{errA}, []shared.FieldError{errB}, []shared.FieldError{errC}))
}

func TestValidateIf(t *testing.T) {
	errs := []shared.FieldError{{Source: "a", Detail: "a"}}

	assert.Equal(t, errs, validateIf(true, errs))
	assert.Nil(t, validateIf(false, errs))
}

func TestValidateIfElse(t *testing.T) {
	errsA := []shared.FieldError{{Source: "a", Detail: "a"}}
	errsB := []shared.FieldError{{Source: "b", Detail: "b"}}

	assert.Equal(t, errsA, validateIfElse(true, errsA, errsB))
	assert.Equal(t, errsB, validateIfElse(false, errsA, errsB))
}

func TestRequired(t *testing.T) {
	assert.Nil(t, required("a", "a"))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "field is required"}}, required("a", ""))
}

func TestValidateDate(t *testing.T) {
	assert.Nil(t, validateDate("a", shared.Date{Time: time.Now()}))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "invalid format"}}, validateDate("a", shared.Date{IsMalformed: true}))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "field is required"}}, validateDate("a", shared.Date{}))
}

func TestValidateAddressEmpty(t *testing.T) {
	address := shared.Address{}
	errors := validateAddress("/test", address)

	assert.Contains(t, errors, shared.FieldError{Source: "/test/line1", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/town", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/country", Detail: "field is required"})
}

func TestValidateAddressValid(t *testing.T) {
	errors := validateAddress("/test", validAddress)

	assert.Empty(t, errors)
}

func TestValidateAddressInvalidCountry(t *testing.T) {
	invalidAddress := shared.Address{
		Line1:   "123 Main St",
		Town:    "Homeland",
		Country: "United Kingdom",
	}
	errors := validateAddress("/test", invalidAddress)

	assert.Contains(t, errors, shared.FieldError{Source: "/test/country", Detail: "must be a valid ISO-3166-1 country code"})
}

type testIsValid string

func (t testIsValid) IsValid() bool { return string(t) == "ok" }

func TestValidateIsValid(t *testing.T) {
	assert.Nil(t, validateIsValid("a", testIsValid("ok")))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "field is required"}}, validateIsValid("a", testIsValid("")))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "invalid value"}}, validateIsValid("a", testIsValid("x")))
}

type testUnset bool

func (t testUnset) Unset() bool { return bool(t) }

func TestValidateUnset(t *testing.T) {
	assert.Nil(t, validateUnset("a", testUnset(true)))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "field must not be provided"}}, validateUnset("a", testUnset(false)))
}

func TestValidateAttorneyEmpty(t *testing.T) {
	attorney := shared.Attorney{}
	errors := validateAttorney("/test", attorney)

	assert.Contains(t, errors, shared.FieldError{Source: "/test/firstNames", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/lastName", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/status", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/dateOfBirth", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/address/line1", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/address/town", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/address/country", Detail: "field is required"})
}

func TestValidateAttorneyValid(t *testing.T) {
	attorney := shared.Attorney{
		Person: shared.Person{
			FirstNames: "Lesia",
			LastName:   "Lathim",
			Address:    validAddress,
		},
		DateOfBirth: newDate("1928-01-18", false),
		Status:      shared.AttorneyStatusActive,
	}
	errors := validateAttorney("/test", attorney)

	assert.Empty(t, errors)
}

func TestValidateAttorneyMalformedDateOfBirth(t *testing.T) {
	attorney := shared.Attorney{
		Person: shared.Person{
			FirstNames: "Lesia",
			LastName:   "Lathim",
			Address:    validAddress,
		},
		DateOfBirth: newDate("bad date", true),
		Status:      shared.AttorneyStatusActive,
	}
	errors := validateAttorney("/test", attorney)

	assert.Contains(t, errors, shared.FieldError{Source: "/test/dateOfBirth", Detail: "invalid format"})
}

func TestValidateAttorneyInvalidStatus(t *testing.T) {
	attorney := shared.Attorney{
		Person: shared.Person{
			FirstNames: "Lesia",
			LastName:   "Lathim",
			Address:    validAddress,
		},
		DateOfBirth: newDate("1928-01-18", false),
		Status:      "bad status",
	}
	errors := validateAttorney("/test", attorney)

	assert.Contains(t, errors, shared.FieldError{Source: "/test/status", Detail: "invalid value"})
}

func TestValidateLpaEmpty(t *testing.T) {
	lpa := shared.LpaInit{}
	errors := Validate(lpa)

	assert.Contains(t, errors, shared.FieldError{Source: "/type", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/donor/firstNames", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/donor/lastName", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/donor/dateOfBirth", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/attorneys", Detail: "at least one attorney is required"})
}

func TestValidateLpaValid(t *testing.T) {
	lpa := shared.LpaInit{
		Type: "hw",
		Donor: shared.Donor{
			Person: shared.Person{
				FirstNames: "Otto",
				LastName:   "Boudreau",
				Address:    validAddress,
			},
			DateOfBirth: newDate("1956-08-08", false),
		},
		Attorneys: []shared.Attorney{
			{
				Person: shared.Person{
					FirstNames: "Sharonda",
					LastName:   "Graciani",
					Address:    validAddress,
				},
				DateOfBirth: newDate("1977-10-30", false),
				Status:      shared.AttorneyStatusActive,
			},
		},
		CertificateProvider: shared.CertificateProvider{
			Person: shared.Person{
				FirstNames: "Some",
				LastName:   "Person",
				Address:    validAddress,
			},
			Email:      "some@example.com",
			CarryOutBy: "online",
		},
		HowAttorneysMakeDecisions:     shared.HowMakeDecisionsJointly,
		LifeSustainingTreatmentOption: shared.LifeSustainingTreatmentOptionA,
		SignedAt:                      time.Now(),
	}
	errors := Validate(lpa)

	assert.Empty(t, errors)
}
