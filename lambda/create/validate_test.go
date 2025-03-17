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

func makeAttorney() shared.Attorney {
	return shared.Attorney{
		Person: shared.Person{
			UID:        "b99af83d-5b6c-44f7-8c03-14004699bdb9",
			FirstNames: "Sharonda",
			LastName:   "Graciani",
		},
		Address:         validAddress,
		AppointmentType: shared.AppointmentTypeOriginal,
		Email:           "some@example.com",
		Channel:         shared.ChannelOnline,
		DateOfBirth:     newDate("1977-10-30"),
		Status:          shared.AttorneyStatusActive,
	}
}

func makeReplacementAttorney() shared.Attorney {
	return shared.Attorney{
		Person: shared.Person{
			UID:        "b99af83d-5b6c-44f7-8c03-14004699bdb9",
			FirstNames: "Sharonda",
			LastName:   "Graciani",
		},
		Address:         validAddress,
		AppointmentType: shared.AppointmentTypeReplacement,
		Email:           "some@example.com",
		Channel:         shared.ChannelOnline,
		DateOfBirth:     newDate("1977-10-30"),
		Status:          shared.AttorneyStatusInactive,
	}
}

func makeCertificateProvider() shared.CertificateProvider {
	return shared.CertificateProvider{
		Person: shared.Person{
			UID:        "b99af83d-5b6c-44f7-8c03-14004699bdb9",
			FirstNames: "Some",
			LastName:   "Person",
		},
		Address: validAddress,
		Email:   "some@example.com",
		Channel: shared.ChannelOnline,
		Phone:   "077777",
	}
}

func makeTrustCorporation() shared.TrustCorporation {
	return shared.TrustCorporation{
		UID:             "af2f7aa6-2f8e-4311-af2a-4855c4686d30",
		Name:            "corp",
		CompanyNumber:   "5",
		Email:           "corp@example.com",
		Address:         validAddress,
		AppointmentType: shared.AppointmentTypeOriginal,
		Status:          shared.AttorneyStatusActive,
		Channel:         shared.ChannelOnline,
	}
}

func makeLpaWithDonorAndActors() shared.LpaInit {
	return shared.LpaInit{
		LpaType:  shared.LpaTypePropertyAndAffairs,
		Channel:  shared.ChannelOnline,
		Language: shared.LangEn,
		Donor: shared.Donor{
			Person: shared.Person{
				UID:        "b99af83d-5b6c-44f7-8c03-14004699bdb9",
				FirstNames: "Otto",
				LastName:   "Boudreau",
			},
			Address:                   validAddress,
			ContactLanguagePreference: shared.LangEn,
			DateOfBirth:               newDate("1956-08-08"),
		},
		CertificateProvider:              makeCertificateProvider(),
		Attorneys:                        []shared.Attorney{makeAttorney()},
		SignedAt:                         time.Now(),
		WitnessedByCertificateProviderAt: time.Now(),
		WhenTheLpaCanBeUsed:              shared.CanUseWhenHasCapacity,
	}
}

func TestValidateAttorneyEmpty(t *testing.T) {
	errors := validateAttorney("/test", shared.Attorney{})

	assert.ElementsMatch(t, []shared.FieldError{
		{Source: "/test/uid", Detail: "field is required"},
		{Source: "/test/firstNames", Detail: "field is required"},
		{Source: "/test/lastName", Detail: "field is required"},
		{Source: "/test/appointmentType", Detail: "field is required"},
		{Source: "/test/status", Detail: "field is required"},
		{Source: "/test/address/line1", Detail: "field is required"},
		{Source: "/test/address/country", Detail: "field is required"},
		{Source: "/test/address/country", Detail: "must be a valid ISO-3166-1 country code"},
		{Source: "/test/channel", Detail: "field is required"},
		{Source: "/test/dateOfBirth", Detail: "field is required"},
	}, errors)
}

func TestValidateAttorneyValid(t *testing.T) {
	errors := validateAttorney("/test", makeAttorney())

	assert.Empty(t, errors)
}

func TestValidateAttorneyMalformedDateOfBirth(t *testing.T) {
	attorney := makeAttorney()
	attorney.DateOfBirth = shared.Date{IsMalformed: true}

	errors := validateAttorney("/test", attorney)

	assert.Equal(t, errors, []shared.FieldError{{Source: "/test/dateOfBirth", Detail: "invalid format"}})
}

func TestValidateAttorneyInvalidStatus(t *testing.T) {
	attorney := makeAttorney()
	attorney.Status = "bad status"

	errors := validateAttorney("/test", attorney)

	assert.Equal(t, errors, []shared.FieldError{{Source: "/test/status", Detail: "invalid value"}})
}

func TestValidateTrustCorporationEmpty(t *testing.T) {
	trustCorporation := shared.TrustCorporation{}
	errors := validateTrustCorporation("/test", trustCorporation)

	assert.ElementsMatch(t, []shared.FieldError{
		{Source: "/test/uid", Detail: "field is required"},
		{Source: "/test/name", Detail: "field is required"},
		{Source: "/test/companyNumber", Detail: "field is required"},
		{Source: "/test/appointmentType", Detail: "field is required"},
		{Source: "/test/status", Detail: "field is required"},
		{Source: "/test/address/line1", Detail: "field is required"},
		{Source: "/test/address/country", Detail: "field is required"},
		{Source: "/test/address/country", Detail: "must be a valid ISO-3166-1 country code"},
		{Source: "/test/channel", Detail: "field is required"},
	}, errors)
}

func TestValidateTrustCorporationValid(t *testing.T) {
	errors := validateTrustCorporation("/test", makeTrustCorporation())

	assert.Empty(t, errors)
}

func TestValidateTrustCorporationInvalidStatus(t *testing.T) {
	trustCorporation := makeTrustCorporation()
	trustCorporation.Status = "bad status"

	errors := validateTrustCorporation("/test", trustCorporation)

	assert.Equal(t, errors, []shared.FieldError{{Source: "/test/status", Detail: "invalid value"}})
}

func TestValidateLpaInvalid(t *testing.T) {
	testcases := map[string]struct {
		lpa            func() shared.LpaInit
		expectedErrors []shared.FieldError
	}{
		"empty": {
			lpa: func() shared.LpaInit { return shared.LpaInit{} },
			expectedErrors: []shared.FieldError{
				{Source: "/lpaType", Detail: "field is required"},
				{Source: "/channel", Detail: "field is required"},
				{Source: "/language", Detail: "field is required"},
				{Source: "/donor/uid", Detail: "field is required"},
				{Source: "/donor/firstNames", Detail: "field is required"},
				{Source: "/donor/lastName", Detail: "field is required"},
				{Source: "/donor/dateOfBirth", Detail: "field is required"},
				{Source: "/donor/address/line1", Detail: "field is required"},
				{Source: "/donor/address/country", Detail: "field is required"},
				{Source: "/donor/address/country", Detail: "must be a valid ISO-3166-1 country code"},
				{Source: "/donor/contactLanguagePreference", Detail: "field is required"},
				{Source: "/certificateProvider/uid", Detail: "field is required"},
				{Source: "/certificateProvider/firstNames", Detail: "field is required"},
				{Source: "/certificateProvider/lastName", Detail: "field is required"},
				{Source: "/certificateProvider/address/line1", Detail: "field is required"},
				{Source: "/certificateProvider/address/country", Detail: "field is required"},
				{Source: "/certificateProvider/address/country", Detail: "must be a valid ISO-3166-1 country code"},
				{Source: "/certificateProvider/channel", Detail: "field is required"},
				{Source: "/certificateProvider/phone", Detail: "field is required"},
				{Source: "/attorneys", Detail: "at least one attorney is required"},
				{Source: "/signedAt", Detail: "field is required"},
				{Source: "/witnessedByCertificateProviderAt", Detail: "field is required"},
			},
		},
		"online certificate provider missing email": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()
				cp := makeCertificateProvider()

				cp.Email = ""
				lpa.CertificateProvider = cp

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/certificateProvider/email", Detail: "field is required"},
			},
		},
		"paper certificate provider with email": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()
				cp := makeCertificateProvider()

				cp.Channel = shared.ChannelPaper
				lpa.CertificateProvider = cp

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/certificateProvider/email", Detail: "field must not be provided"},
			},
		},
		"single attorney with decisions": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()
				lpa.HowAttorneysMakeDecisions = shared.HowMakeDecisionsJointly

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howAttorneysMakeDecisions", Detail: "field must not be provided"},
			},
		},
		"multiple attorneys without decisions": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()

				lpa.Attorneys = []shared.Attorney{makeAttorney(), makeAttorney()}

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howAttorneysMakeDecisions", Detail: "field is required"},
			},
		},
		"multiple attorneys mixed without details": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()

				lpa.Attorneys = []shared.Attorney{makeAttorney(), makeAttorney()}
				lpa.HowAttorneysMakeDecisions = shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howAttorneysMakeDecisionsDetails", Detail: "field is required"},
			},
		},
		"multiple attorneys not mixed with details": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()

				lpa.Attorneys = []shared.Attorney{makeAttorney(), makeAttorney()}
				lpa.HowAttorneysMakeDecisions = shared.HowMakeDecisionsJointly
				lpa.HowAttorneysMakeDecisionsDetails = "something"

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howAttorneysMakeDecisionsDetails", Detail: "field must not be provided"},
			},
		},
		"online attorney missing email": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()
				lpa.Attorneys[0].Email = ""

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/attorneys/0/email", Detail: "field is required"},
			},
		},
		"paper attorney with email": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()
				a := makeAttorney()

				a.Channel = shared.ChannelPaper
				lpa.Attorneys = []shared.Attorney{a}

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/attorneys/0/email", Detail: "field must not be provided"},
			},
		},
		"single replacement attorney with decisions": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()

				lpa.Attorneys = []shared.Attorney{makeReplacementAttorney()}
				lpa.HowReplacementAttorneysMakeDecisions = shared.HowMakeDecisionsJointly

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howReplacementAttorneysMakeDecisions", Detail: "field must not be provided"},
			},
		},
		"multiple replacement attorneys without decisions": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()

				lpa.Attorneys = []shared.Attorney{makeReplacementAttorney(), makeReplacementAttorney()}

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howReplacementAttorneysMakeDecisions", Detail: "field is required"},
			},
		},
		"attorneys jointly and severally multiple replacement attorneys without step in": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()

				lpa.Attorneys = []shared.Attorney{makeReplacementAttorney(), makeReplacementAttorney(), makeAttorney(), makeAttorney()}
				lpa.HowAttorneysMakeDecisions = shared.HowMakeDecisionsJointlyAndSeverally

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howReplacementAttorneysStepIn", Detail: "field is required"},
			},
		},
		"attorneys jointly and severally multiple replacement attorneys with step in another way no details": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()

				lpa.Attorneys = []shared.Attorney{makeReplacementAttorney(), makeReplacementAttorney(), makeAttorney(), makeAttorney()}
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
				lpa := makeLpaWithDonorAndActors()

				lpa.Attorneys = []shared.Attorney{makeReplacementAttorney(), makeReplacementAttorney(), makeAttorney(), makeAttorney()}
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
				lpa := makeLpaWithDonorAndActors()

				lpa.Attorneys = []shared.Attorney{makeReplacementAttorney(), makeReplacementAttorney(), makeAttorney(), makeAttorney()}
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
				lpa := makeLpaWithDonorAndActors()
				lpa.Attorneys = []shared.Attorney{makeReplacementAttorney(), makeReplacementAttorney()}
				lpa.HowReplacementAttorneysMakeDecisions = shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/howReplacementAttorneysMakeDecisionsDetails", Detail: "field is required"},
			},
		},
		"multiple replacement attorneys not mixed with details": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()

				lpa.Attorneys = []shared.Attorney{makeReplacementAttorney(), makeReplacementAttorney()}
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
				lpa := makeLpaWithDonorAndActors()

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
				lpa := makeLpaWithDonorAndActors()

				lpa.WhenTheLpaCanBeUsed = shared.CanUseUnset
				lpa.LifeSustainingTreatmentOption = shared.LifeSustainingTreatmentOptionA

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/whenTheLpaCanBeUsed", Detail: "field is required"},
				{Source: "/lifeSustainingTreatmentOption", Detail: "field must not be provided"},
			},
		},
		"online trust corporation missing email": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()
				tc := makeTrustCorporation()

				tc.Email = ""
				lpa.TrustCorporations = []shared.TrustCorporation{tc}
				lpa.HowAttorneysMakeDecisions = shared.HowMakeDecisionsJointly

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/trustCorporations/0/email", Detail: "field is required"},
			},
		},
		"paper trust corporation with email": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()
				tc := makeTrustCorporation()

				tc.Channel = shared.ChannelPaper
				lpa.TrustCorporations = []shared.TrustCorporation{tc}
				lpa.HowAttorneysMakeDecisions = shared.HowMakeDecisionsJointly

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/trustCorporations/0/email", Detail: "field must not be provided"},
			},
		},
		"incorrect identity check": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()
				lpa.Donor.IdentityCheck = &shared.IdentityCheck{Type: "what"}

				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/donor/identityCheck/checkedAt", Detail: "field is required"},
				{Source: "/donor/identityCheck/type", Detail: "invalid value"},
			},
		},
		"independent witness missing witnessed at": {
			lpa: func() shared.LpaInit {
				lpa := makeLpaWithDonorAndActors()
				lpa.IndependentWitness = &shared.IndependentWitness{
					Person: shared.Person{
						UID:        "b99af83d-5b6c-44f7-8c03-14004699bdb9",
						FirstNames: "Some",
						LastName:   "Person",
					},
					Address: validAddress,
					Phone:   "077777",
				}
				return lpa
			},
			expectedErrors: []shared.FieldError{
				{Source: "/witnessedByIndependentWitnessAt", Detail: "field is required"},
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
	lpa := makeLpaWithDonorAndActors()
	lpa.TrustCorporations = []shared.TrustCorporation{makeTrustCorporation()}
	lpa.HowAttorneysMakeDecisions = shared.HowMakeDecisionsJointly

	errors := Validate(lpa)

	assert.Empty(t, errors)
}
