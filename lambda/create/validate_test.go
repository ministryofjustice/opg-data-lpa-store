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
	attorney := shared.Attorney{
		Person: shared.Person{
			FirstNames: "Sharonda",
			LastName:   "Graciani",
			Address:    validAddress,
		},
		DateOfBirth: newDate("1977-10-30", false),
		Status:      shared.AttorneyStatusActive,
	}

	replacementAttorney := shared.Attorney{
		Person: shared.Person{
			FirstNames: "Sharonda",
			LastName:   "Graciani",
			Address:    validAddress,
		},
		DateOfBirth: newDate("1977-10-30", false),
		Status:      shared.AttorneyStatusReplacement,
	}

	certificateProvider := shared.CertificateProvider{
		Person: shared.Person{
			FirstNames: "Some",
			LastName:   "Person",
			Address:    validAddress,
		},
		Email:   "some@example.com",
		Channel: shared.ChannelOnline,
	}

	lpaWithDonorAndActors := shared.LpaInit{
		LpaType: shared.LpaTypePropertyAndAffairs,
		Channel: shared.ChannelOnline,
		Donor: shared.Donor{
			Person: shared.Person{
				FirstNames: "Otto",
				LastName:   "Boudreau",
				Address:    validAddress,
			},
			DateOfBirth: newDate("1956-08-08", false),
		},
		CertificateProvider: certificateProvider,
		Attorneys:           []shared.Attorney{attorney},
		SignedAt:            time.Now(),
		WhenTheLpaCanBeUsed: shared.CanUseWhenHasCapacity,
	}

	testcases := map[string]struct {
		lpa            func() shared.LpaInit
		expectedErrors []shared.FieldError
	}{
		"empty": {
			lpa: func() shared.LpaInit { return shared.LpaInit{} },
			expectedErrors: []shared.FieldError{
				{Source: "/lpaType", Detail: "field is required"},
				{Source: "/channel", Detail: "field is required"},
				{Source: "/donor/firstNames", Detail: "field is required"},
				{Source: "/donor/lastName", Detail: "field is required"},
				{Source: "/donor/dateOfBirth", Detail: "field is required"},
				{Source: "/donor/address/line1", Detail: "field is required"},
				{Source: "/donor/address/town", Detail: "field is required"},
				{Source: "/donor/address/country", Detail: "field is required"},
				{Source: "/donor/address/country", Detail: "must be a valid ISO-3166-1 country code"},
				{Source: "/certificateProvider/firstNames", Detail: "field is required"},
				{Source: "/certificateProvider/lastName", Detail: "field is required"},
				{Source: "/certificateProvider/address/line1", Detail: "field is required"},
				{Source: "/certificateProvider/address/town", Detail: "field is required"},
				{Source: "/certificateProvider/address/country", Detail: "field is required"},
				{Source: "/certificateProvider/address/country", Detail: "must be a valid ISO-3166-1 country code"},
				{Source: "/certificateProvider/channel", Detail: "field is required"},
				{Source: "/attorneys", Detail: "at least one attorney is required"},
				{Source: "/signedAt", Detail: "field is required"},
			},
		},
		"online certificate provider missing email": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.CertificateProvider.Email = ""

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/certificateProvider/email", Detail: "field is required"},
			},
		},
		"paper certificate provider with email": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.Attorneys = []shared.Attorney{attorney}
				lpa.CertificateProvider.Channel = shared.ChannelPaper

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/certificateProvider/email", Detail: "field must not be provided"},
			},
		},
		"single attorney with decisions": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.HowAttorneysMakeDecisions = shared.HowMakeDecisionsJointly

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howAttorneysMakeDecisions", Detail: "field must not be provided"},
			},
		},
		"multiple attorneys without decisions": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.Attorneys = []shared.Attorney{attorney, attorney}

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howAttorneysMakeDecisions", Detail: "field is required"},
			},
		},
		"multiple attorneys mixed without details": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.Attorneys = []shared.Attorney{attorney, attorney}
				lpa.HowAttorneysMakeDecisions = shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howAttorneysMakeDecisionsDetails", Detail: "field is required"},
			},
		},
		"multiple attorneys not mixed with details": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.Attorneys = []shared.Attorney{attorney, attorney}
				lpa.HowAttorneysMakeDecisions = shared.HowMakeDecisionsJointly
				lpa.HowAttorneysMakeDecisionsDetails = "something"

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howAttorneysMakeDecisionsDetails", Detail: "field must not be provided"},
			},
		},
		"single replacement attorney with decisions": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.Attorneys = []shared.Attorney{replacementAttorney}
				lpa.HowReplacementAttorneysMakeDecisions = shared.HowMakeDecisionsJointly

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howReplacementAttorneysMakeDecisions", Detail: "field must not be provided"},
			},
		},
		"multiple replacement attorneys without decisions": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.Attorneys = []shared.Attorney{replacementAttorney, replacementAttorney}

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howReplacementAttorneysMakeDecisions", Detail: "field is required"},
			},
		},
		"attorneys jointly and severally multiple replacement attorneys without step in": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.Attorneys = []shared.Attorney{replacementAttorney, replacementAttorney, attorney, attorney}
				lpa.HowAttorneysMakeDecisions = shared.HowMakeDecisionsJointlyAndSeverally

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howReplacementAttorneysStepIn", Detail: "field is required"},
			},
		},
		"attorneys jointly and severally multiple replacement attorneys with step in another way no details": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.Attorneys = []shared.Attorney{replacementAttorney, replacementAttorney, attorney, attorney}
				lpa.HowAttorneysMakeDecisions = shared.HowMakeDecisionsJointlyAndSeverally
				lpa.HowReplacementAttorneysStepIn = shared.HowStepInAnotherWay

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howReplacementAttorneysStepInDetails", Detail: "field is required"},
			},
		},
		"attorneys jointly and severally multiple replacement attorneys with step in all no decisions": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.Attorneys = []shared.Attorney{replacementAttorney, replacementAttorney, attorney, attorney}
				lpa.HowAttorneysMakeDecisions = shared.HowMakeDecisionsJointlyAndSeverally
				lpa.HowReplacementAttorneysStepIn = shared.HowStepInAllCanNoLongerAct

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howReplacementAttorneysMakeDecisions", Detail: "field is required"},
			},
		},
		"attorneys jointly and severally multiple replacement attorneys with step in one with decisions": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.Attorneys = []shared.Attorney{replacementAttorney, replacementAttorney, attorney, attorney}
				lpa.HowAttorneysMakeDecisions = shared.HowMakeDecisionsJointlyAndSeverally
				lpa.HowReplacementAttorneysStepIn = shared.HowStepInOneCanNoLongerAct
				lpa.HowReplacementAttorneysMakeDecisions = shared.HowMakeDecisionsJointly

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howReplacementAttorneysMakeDecisions", Detail: "field must not be provided"},
			},
		},
		"multiple replacement attorneys mixed without details": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.Attorneys = []shared.Attorney{replacementAttorney, replacementAttorney}
				lpa.HowReplacementAttorneysMakeDecisions = shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howReplacementAttorneysMakeDecisionsDetails", Detail: "field is required"},
			},
		},
		"multiple replacement attorneys not mixed with details": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.Attorneys = []shared.Attorney{replacementAttorney, replacementAttorney}
				lpa.HowReplacementAttorneysMakeDecisions = shared.HowMakeDecisionsJointly
				lpa.HowReplacementAttorneysMakeDecisionsDetails = "something"

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howReplacementAttorneysMakeDecisionsDetails", Detail: "field must not be provided"},
			},
		},
		"health welfare with when can be used": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.LpaType = shared.LpaTypePersonalWelfare
				lpa.WhenTheLpaCanBeUsed = shared.CanUseWhenHasCapacity

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/lifeSustainingTreatmentOption", Detail: "field is required"},
				{Source: "/whenTheLpaCanBeUsed", Detail: "field must not be provided"},
			},
		},
		"property finance with life sustaining treatment": {
			lpa: func() shared.LpaInit {
				lpa := lpaWithDonorAndActors
				lpa.WhenTheLpaCanBeUsed = shared.CanUseUnset
				lpa.LifeSustainingTreatmentOption = shared.LifeSustainingTreatmentOptionA

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/whenTheLpaCanBeUsed", Detail: "field is required"},
				{Source: "/lifeSustainingTreatmentOption", Detail: "field must not be provided"},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedErrors, Validate(tc.lpa()))
		})
	}
}

func TestValidateLpaValid(t *testing.T) {
	lpa := shared.LpaInit{
		LpaType: shared.LpaTypePersonalWelfare,
		Channel: shared.ChannelOnline,
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
			Email:   "some@example.com",
			Channel: shared.ChannelOnline,
		},
		LifeSustainingTreatmentOption: shared.LifeSustainingTreatmentOptionA,
		SignedAt:                      time.Now(),
	}
	errors := Validate(lpa)

	assert.Empty(t, errors)
}
