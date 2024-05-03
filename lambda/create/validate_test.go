package main

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

var validAddress = shared.Address{
	Line1:   "123 Main St",
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
	assert.Contains(t, errors, shared.FieldError{Source: "/test/address/country", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/channel", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/uid", Detail: "field is required"})
}

func TestValidateAttorneyValid(t *testing.T) {
	attorney := shared.Attorney{
		Person: shared.Person{
			UID:        "0a266ff6-1c7b-49b7-acd0-047f1dcda2ce",
			FirstNames: "Lesia",
			LastName:   "Lathim",
			Address:    validAddress,
		},
		DateOfBirth: newDate("1928-01-18"),
		Status:      shared.AttorneyStatusActive,
		Channel:     shared.ChannelOnline,
		Email:       "a@example.com",
	}
	errors := validateAttorney("/test", attorney)

	assert.Empty(t, errors)
}

func TestValidateAttorneyMalformedDateOfBirth(t *testing.T) {
	attorney := shared.Attorney{
		Person: shared.Person{
			UID:        "0a266ff6-1c7b-49b7-acd0-047f1dcda2ce",
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
			UID:        "0a266ff6-1c7b-49b7-acd0-047f1dcda2ce",
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
	assert.Contains(t, errors, shared.FieldError{Source: "/test/status", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/address/line1", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/address/country", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/channel", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/uid", Detail: "field is required"})
}

func TestValidateTrustCorporationValid(t *testing.T) {
	trustCorporation := shared.TrustCorporation{
		UID:           "af2f7aa6-2f8e-4311-af2a-4855c4686d30",
		Name:          "corp",
		CompanyNumber: "5",
		Email:         "corp@example.com",
		Address:       validAddress,
		Status:        shared.AttorneyStatusActive,
		Channel:       shared.ChannelOnline,
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
				{Source: "/channel", Detail: "field is required"},
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
		"online attorney missing email": {
			lpa: shared.LpaInit{
				Attorneys: []shared.Attorney{
					{
						Channel: shared.ChannelOnline,
						Status:  shared.AttorneyStatusActive,
					},
				},
			},
			contains: []shared.FieldError{
				{Source: "/attorneys/0/email", Detail: "field is required"},
			},
		},
		"paper attorney with email": {
			lpa: shared.LpaInit{
				Attorneys: []shared.Attorney{
					{
						Channel: shared.ChannelPaper,
						Email:   "a@example.com",
						Status:  shared.AttorneyStatusActive,
					},
				},
			},
			contains: []shared.FieldError{
				{Source: "/attorneys/0/email", Detail: "field must not be provided"},
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
		"online trust corporation missing email": {
			lpa: shared.LpaInit{
				TrustCorporations: []shared.TrustCorporation{
					{
						Channel: shared.ChannelOnline,
					},
				},
			},
			contains: []shared.FieldError{
				{Source: "/trustCorporations/0/email", Detail: "field is required"},
			},
		},
		"paper trust corporation with email": {
			lpa: shared.LpaInit{
				TrustCorporations: []shared.TrustCorporation{
					{
						Channel: shared.ChannelPaper,
						Email:   "a@example.com",
					},
				},
			},
			contains: []shared.FieldError{
				{Source: "/trustCorporations/0/email", Detail: "field must not be provided"},
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
		Channel: shared.ChannelOnline,
		Donor: shared.Donor{
			Person: shared.Person{
				UID:        "e0f311fd-fe38-40c9-8e82-8263d0f578d9",
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
					UID:        "b99af83d-5b6c-44f7-8c03-14004699bdb9",
					FirstNames: "Sharonda",
					LastName:   "Graciani",
					Address:    validAddress,
				},
				DateOfBirth: newDate("1977-10-30"),
				Status:      shared.AttorneyStatusActive,
				Channel:     shared.ChannelOnline,
				Email:       "a@example.com",
			},
		},
		CertificateProvider: shared.CertificateProvider{
			Person: shared.Person{
				UID:        "613d3e2c-4091-42d6-97b3-21bd76f4ffed",
				FirstNames: "Some",
				LastName:   "Person",
				Address:    validAddress,
			},
			Phone:   "070009000",
			Email:   "some@example.com",
			Channel: shared.ChannelOnline,
		},
		TrustCorporations: []shared.TrustCorporation{
			{
				UID:           "af2f7aa6-2f8e-4311-af2a-4855c4686d30",
				Name:          "corp",
				CompanyNumber: "5",
				Email:         "corp@example.com",
				Address:       validAddress,
				Status:        shared.AttorneyStatusActive,
				Channel:       shared.ChannelOnline,
			},
		},
		LifeSustainingTreatmentOption: shared.LifeSustainingTreatmentOptionA,
		SignedAt:                      time.Now(),
		HowAttorneysMakeDecisions:     shared.HowMakeDecisionsJointly,
	}
	errors := Validate(lpa)

	assert.Empty(t, errors)
}
