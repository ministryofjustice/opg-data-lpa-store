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

func TestEmpty(t *testing.T) {
	assert.Nil(t, empty("a", ""))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "field must not be provided"}}, empty("a", "a"))
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

func TestValidateLpaInvalid(t *testing.T) {
	testcases := map[string]struct {
		lpa      shared.LpaInit
		contains []shared.FieldError
	}{
		"empty": {
			contains: []shared.FieldError{
				{Source: "/type", Detail: "field is required"},
				{Source: "/donor/firstNames", Detail: "field is required"},
				{Source: "/donor/lastName", Detail: "field is required"},
				{Source: "/donor/dateOfBirth", Detail: "field is required"},
				{Source: "/attorneys", Detail: "at least one attorney is required"},
			},
		},
		"online certificate provider missing email": {
			lpa: shared.LpaInit{
				CertificateProvider: shared.CertificateProvider{
					CarryOutBy: shared.CarryOutByOnline,
				},
			},
			contains: []shared.FieldError{
				{Source: "/certificateProvider/email", Detail: "field is required"},
			},
		},
		"paper certificate provider with email": {
			lpa: shared.LpaInit{
				CertificateProvider: shared.CertificateProvider{
					CarryOutBy: shared.CarryOutByPaper,
					Email:      "something",
				},
			},
			contains: []shared.FieldError{
				{Source: "/certificateProvider/email", Detail: "field must not be provided"},
			},
		},
		"single attorney with decisions": {
			lpa: shared.LpaInit{
				Attorneys:                 []shared.Attorney{{Status: shared.AttorneyStatusActive}},
				HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointly,
			},
			contains: []shared.FieldError{
				{Source: "/howAttorneysMakeDecisions", Detail: "field must not be provided"},
			},
		},
		"multiple attorneys without decisions": {
			lpa: shared.LpaInit{
				Attorneys: []shared.Attorney{{Status: shared.AttorneyStatusActive}, {Status: shared.AttorneyStatusActive}},
			},
			contains: []shared.FieldError{
				{Source: "/howAttorneysMakeDecisions", Detail: "field is required"},
			},
		},
		"multiple attorneys mixed without details": {
			lpa: shared.LpaInit{
				Attorneys:                 []shared.Attorney{{Status: shared.AttorneyStatusActive}, {Status: shared.AttorneyStatusActive}},
				HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
			},
			contains: []shared.FieldError{
				{Source: "/howAttorneysMakeDecisionsDetails", Detail: "field is required"},
			},
		},
		"multiple attorneys not mixed with details": {
			lpa: shared.LpaInit{
				Attorneys:                        []shared.Attorney{{Status: shared.AttorneyStatusActive}, {Status: shared.AttorneyStatusActive}},
				HowAttorneysMakeDecisions:        shared.HowMakeDecisionsJointly,
				HowAttorneysMakeDecisionsDetails: "something",
			},
			contains: []shared.FieldError{
				{Source: "/howAttorneysMakeDecisionsDetails", Detail: "field must not be provided"},
			},
		},
		"single replacement attorney with decisions": {
			lpa: shared.LpaInit{
				Attorneys:                            []shared.Attorney{{Status: shared.AttorneyStatusReplacement}},
				HowReplacementAttorneysMakeDecisions: shared.HowMakeDecisionsJointly,
			},
			contains: []shared.FieldError{
				{Source: "/howReplacementAttorneysMakeDecisions", Detail: "field must not be provided"},
			},
		},
		"multiple replacement attorneys without decisions": {
			lpa: shared.LpaInit{
				Attorneys: []shared.Attorney{{Status: shared.AttorneyStatusReplacement}, {Status: shared.AttorneyStatusReplacement}},
			},
			contains: []shared.FieldError{
				{Source: "/howReplacementAttorneysMakeDecisions", Detail: "field is required"},
			},
		},
		"attorneys jointly and severally multiple replacement attorneys without step in": {
			lpa: shared.LpaInit{
				Attorneys:                 []shared.Attorney{{Status: shared.AttorneyStatusReplacement}, {Status: shared.AttorneyStatusReplacement}},
				HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
			},
			contains: []shared.FieldError{
				{Source: "/howReplacementAttorneysStepIn", Detail: "field is required"},
			},
		},
		"attorneys jointly and severally multiple replacement attorneys with step in another way no details": {
			lpa: shared.LpaInit{
				Attorneys:                     []shared.Attorney{{Status: shared.AttorneyStatusReplacement}, {Status: shared.AttorneyStatusReplacement}},
				HowAttorneysMakeDecisions:     shared.HowMakeDecisionsJointlyAndSeverally,
				HowReplacementAttorneysStepIn: shared.HowStepInAnotherWay,
			},
			contains: []shared.FieldError{
				{Source: "/howReplacementAttorneysStepInDetails", Detail: "field is required"},
			},
		},
		"attorneys jointly and severally multiple replacement attorneys with step in all no decisions": {
			lpa: shared.LpaInit{
				Attorneys:                     []shared.Attorney{{Status: shared.AttorneyStatusReplacement}, {Status: shared.AttorneyStatusReplacement}},
				HowAttorneysMakeDecisions:     shared.HowMakeDecisionsJointlyAndSeverally,
				HowReplacementAttorneysStepIn: shared.HowStepInAllCanNoLongerAct,
			},
			contains: []shared.FieldError{
				{Source: "/howReplacementAttorneysMakeDecisions", Detail: "field is required"},
			},
		},
		"attorneys jointly and severally multiple replacement attorneys with step in one with decisions": {
			lpa: shared.LpaInit{
				Attorneys:                            []shared.Attorney{{Status: shared.AttorneyStatusReplacement}, {Status: shared.AttorneyStatusReplacement}},
				HowAttorneysMakeDecisions:            shared.HowMakeDecisionsJointlyAndSeverally,
				HowReplacementAttorneysStepIn:        shared.HowStepInOneCanNoLongerAct,
				HowReplacementAttorneysMakeDecisions: shared.HowMakeDecisionsJointly,
			},
			contains: []shared.FieldError{
				{Source: "/howReplacementAttorneysMakeDecisions", Detail: "field must not be provided"},
			},
		},
		"multiple replacement attorneys mixed without details": {
			lpa: shared.LpaInit{
				Attorneys:                            []shared.Attorney{{Status: shared.AttorneyStatusReplacement}, {Status: shared.AttorneyStatusReplacement}},
				HowReplacementAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
			},
			contains: []shared.FieldError{
				{Source: "/howReplacementAttorneysMakeDecisionsDetails", Detail: "field is required"},
			},
		},
		"multiple replacement attorneys not mixed with details": {
			lpa: shared.LpaInit{
				Attorneys:                                   []shared.Attorney{{Status: shared.AttorneyStatusReplacement}, {Status: shared.AttorneyStatusReplacement}},
				HowReplacementAttorneysMakeDecisions:        shared.HowMakeDecisionsJointly,
				HowReplacementAttorneysMakeDecisionsDetails: "something",
			},
			contains: []shared.FieldError{
				{Source: "/howReplacementAttorneysMakeDecisionsDetails", Detail: "field must not be provided"},
			},
		},
		"health welfare with when can be used": {
			lpa: shared.LpaInit{
				Type:                shared.TypeHealthWelfare,
				WhenTheLpaCanBeUsed: shared.CanUseWhenHasCapacity,
			},
			contains: []shared.FieldError{
				{Source: "/whenTheLpaCanBeUsed", Detail: "field must not be provided"},
				{Source: "/lifeSustainingTreatmentOption", Detail: "field is required"},
			},
		},
		"property finance with life sustaining treatment": {
			lpa: shared.LpaInit{
				Type:                          shared.TypePropertyFinance,
				LifeSustainingTreatmentOption: shared.LifeSustainingTreatmentOptionA,
			},
			contains: []shared.FieldError{
				{Source: "/whenTheLpaCanBeUsed", Detail: "field is required"},
				{Source: "/lifeSustainingTreatmentOption", Detail: "field must not be provided"},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			errors := Validate(tc.lpa)
			for _, e := range tc.contains {
				assert.Contains(t, errors, e)
			}
		})
	}
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
		LifeSustainingTreatmentOption: shared.LifeSustainingTreatmentOptionA,
		SignedAt:                      time.Now(),
	}
	errors := Validate(lpa)

	assert.Empty(t, errors)
}
