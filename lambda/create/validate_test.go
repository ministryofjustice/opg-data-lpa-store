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

func newDate(date string) shared.Date {
	d := shared.Date{}
	_ = d.UnmarshalText([]byte(date))
	return d
}

func TestCountAttorneys(t *testing.T) {
	actives, replacements := countAttorneys([]shared.Attorney{}, []shared.TrustCorporation{})
	assert.Equal(t, 0, actives)
	assert.Equal(t, 0, replacements)

	actives, replacements = countAttorneys([]shared.Attorney{
		{Status: shared.AttorneyStatusReplacement},
		{Status: shared.AttorneyStatusActive},
		{Status: shared.AttorneyStatusReplacement},
	}, []shared.TrustCorporation{
		{Status: shared.AttorneyStatusReplacement},
		{Status: shared.AttorneyStatusActive},
	})
	assert.Equal(t, 2, actives)
	assert.Equal(t, 3, replacements)
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
		DateOfBirth: newDate("1928-01-18"),
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
		DateOfBirth: shared.Date{IsMalformed: true},
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
		DateOfBirth: newDate("1928-01-18"),
		Status:      "bad status",
	}
	errors := validateAttorney("/test", attorney)

	assert.Contains(t, errors, shared.FieldError{Source: "/test/status", Detail: "invalid value"})
}

func TestValidateTrustCorporationEmpty(t *testing.T) {
	trustCorporation := shared.TrustCorporation{}
	errors := validateTrustCorporation("/test", trustCorporation)

	assert.Contains(t, errors, shared.FieldError{Source: "/test/name", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/companyNumber", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/email", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/status", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/address/line1", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/address/town", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/address/country", Detail: "field is required"})
}

func TestValidateTrustCorporationValid(t *testing.T) {
	trustCorporation := shared.TrustCorporation{
		Name:          "corp",
		CompanyNumber: "5",
		Email:         "corp@example.com",
		Address:       validAddress,
		Status:        shared.AttorneyStatusActive,
	}
	errors := validateTrustCorporation("/test", trustCorporation)

	assert.Empty(t, errors)
}

func TestValidateTrustCorporationInvalidStatus(t *testing.T) {
	trustCorporation := shared.TrustCorporation{
		Status: "bad status",
	}
	errors := validateTrustCorporation("/test", trustCorporation)

	assert.Contains(t, errors, shared.FieldError{Source: "/test/status", Detail: "invalid value"})
}

func TestValidateLpaInvalid(t *testing.T) {
	testcases := map[string]struct {
		lpa      shared.LpaInit
		contains []shared.FieldError
	}{
		"empty": {
			contains: []shared.FieldError{
				{Source: "/lpaType", Detail: "field is required"},
				{Source: "/donor/firstNames", Detail: "field is required"},
				{Source: "/donor/lastName", Detail: "field is required"},
				{Source: "/donor/dateOfBirth", Detail: "field is required"},
				{Source: "/donor/contactLanguagePreference", Detail: "field is required"},
				{Source: "/attorneys", Detail: "at least one attorney is required"},
				{Source: "/certificateProvider/phone", Detail: "field is required"},
			},
		},
		"online certificate provider missing email": {
			lpa: shared.LpaInit{
				CertificateProvider: shared.CertificateProvider{
					Channel: shared.ChannelOnline,
				},
			},
			contains: []shared.FieldError{
				{Source: "/certificateProvider/email", Detail: "field is required"},
			},
		},
		"paper certificate provider with email": {
			lpa: shared.LpaInit{
				CertificateProvider: shared.CertificateProvider{
					Channel: shared.ChannelPaper,
					Email:   "something",
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
				LpaType:             shared.LpaTypePersonalWelfare,
				WhenTheLpaCanBeUsed: shared.CanUseWhenHasCapacity,
			},
			contains: []shared.FieldError{
				{Source: "/whenTheLpaCanBeUsed", Detail: "field must not be provided"},
				{Source: "/lifeSustainingTreatmentOption", Detail: "field is required"},
			},
		},
		"property finance with life sustaining treatment": {
			lpa: shared.LpaInit{
				LpaType:                       shared.LpaTypePropertyAndAffairs,
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
		LpaType: shared.LpaTypePersonalWelfare,
		Donor: shared.Donor{
			Person: shared.Person{
				FirstNames: "Otto",
				LastName:   "Boudreau",
				Address:    validAddress,
			},
			DateOfBirth:               newDate("1956-08-08"),
			ContactLanguagePreference: shared.LangEn,
		},
		Attorneys: []shared.Attorney{
			{
				Person: shared.Person{
					FirstNames: "Sharonda",
					LastName:   "Graciani",
					Address:    validAddress,
				},
				DateOfBirth: newDate("1977-10-30"),
				Status:      shared.AttorneyStatusActive,
			},
		},
		CertificateProvider: shared.CertificateProvider{
			Person: shared.Person{
				FirstNames: "Some",
				LastName:   "Person",
				Address:    validAddress,
			},
			Phone:   "070009000",
			Email:   "some@example.com",
			Channel: shared.ChannelOnline,
		},
		LifeSustainingTreatmentOption: shared.LifeSustainingTreatmentOptionA,
		SignedAt:                      time.Now(),
	}
	errors := Validate(lpa)

	assert.Empty(t, errors)
}
