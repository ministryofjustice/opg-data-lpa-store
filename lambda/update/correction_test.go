package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func createDate(date string) shared.Date {
	d := shared.Date{}
	_ = d.UnmarshalText([]byte(date))
	return d
}

func ptrTo[T any](v T) *T {
	return &v
}

func TestDonorCorrectionApply(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	lpa := &shared.Lpa{
		LpaInit: shared.LpaInit{
			Donor: shared.Donor{
				Person: shared.Person{
					FirstNames: "donor-firstname",
					LastName:   "donor-lastname",
				},
				OtherNamesKnownBy: "donor-otherNames",
				DateOfBirth:       createDate("1990-01-02"),
				Address: shared.Address{
					Line1:    "123 Main St",
					Town:     "Anytown",
					Postcode: "A11 B22",
					Country:  "IE",
				},
				Email: "john.doe@example.com",
			},
			SignedAt:  yesterday,
			Attorneys: []shared.Attorney{},
		},
	}

	correction := Correction{
		Donor: DonorCorrection{
			FirstNames:        "Jane",
			LastName:          "Smith",
			OtherNamesKnownBy: "Janey",
			DateOfBirth:       createDate("2000-11-11"),
			Address: shared.Address{
				Line1:    "456 Another St",
				Town:     "Othertown",
				Postcode: "B22 A11",
				Country:  "IE",
			},
			Email: "jane.smith@example.com",
		},
		SignedAt: now,
	}

	errors := correction.Apply(lpa)

	assert.Empty(t, errors)
	assert.Equal(t, correction.Donor.FirstNames, lpa.Donor.FirstNames)
	assert.Equal(t, correction.Donor.LastName, lpa.Donor.LastName)
	assert.Equal(t, correction.Donor.OtherNamesKnownBy, lpa.Donor.OtherNamesKnownBy)
	assert.Equal(t, correction.Donor.DateOfBirth, lpa.Donor.DateOfBirth)
	assert.Equal(t, correction.Donor.Address, lpa.Donor.Address)
	assert.Equal(t, correction.Donor.Email, lpa.Donor.Email)
	assert.Equal(t, correction.SignedAt, lpa.SignedAt)
}

func TestAttorneyCorrectionApply(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	lpa := &shared.Lpa{
		LpaInit: shared.LpaInit{
			Attorneys: []shared.Attorney{
				shared.Attorney{},
				{
					Person: shared.Person{
						FirstNames: "attorney-firstname",
						LastName:   "attorney-lastname",
					},
					DateOfBirth: createDate("1990-01-02"),
					Address: shared.Address{
						Line1:    "123 Main St",
						Town:     "Anytown",
						Postcode: "A11 B22",
						Country:  "IE",
					},
					Email:    "test@test.com",
					Mobile:   "0123456789",
					SignedAt: &yesterday,
				},
			},
		},
	}

	index := 1
	correction := Correction{
		Attorney: AttorneyCorrection{
			Index:       &index,
			FirstNames:  "Jane",
			LastName:    "Smith",
			DateOfBirth: createDate("2000-11-11"),
			Address: shared.Address{
				Line1:    "456 Another St",
				Town:     "Othertown",
				Postcode: "B22 A11",
				Country:  "GB",
			},
			Email:    "jane.smith@example.com",
			Mobile:   "987654321",
			SignedAt: now,
		},
	}

	errors := correction.Apply(lpa)

	assert.Empty(t, errors)
	assert.Equal(t, correction.Attorney.FirstNames, lpa.Attorneys[index].FirstNames)
	assert.Equal(t, correction.Attorney.LastName, lpa.Attorneys[index].LastName)
	assert.Equal(t, correction.Attorney.DateOfBirth, lpa.Attorneys[index].DateOfBirth)
	assert.Equal(t, correction.Attorney.Address, lpa.Attorneys[index].Address)
	assert.Equal(t, correction.Attorney.Email, lpa.Attorneys[index].Email)
	assert.Equal(t, correction.Attorney.Mobile, lpa.Attorneys[index].Mobile)
	assert.Equal(t, correction.Attorney.SignedAt, *lpa.Attorneys[index].SignedAt)
}

func TestCorrectionRegisteredLpa(t *testing.T) {
	lpa := &shared.Lpa{
		Status: shared.LpaStatusRegistered,
		LpaInit: shared.LpaInit{
			Channel: "paper",
			Donor: shared.Donor{
				Person: shared.Person{
					FirstNames: "donor-firstname",
				},
			},
		},
	}

	correction := Correction{
		Donor: DonorCorrection{
			FirstNames: "Jane",
		},
	}
	errors := correction.Apply(lpa)

	assert.Equal(t, errors, []shared.FieldError{{Source: "/type", Detail: "Cannot make corrections to a Registered LPA"}})
}

func TestCorrectionLpaSignedOnlineChannel(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	lpa := &shared.Lpa{
		LpaInit: shared.LpaInit{
			Channel:  "online",
			SignedAt: yesterday,
		},
	}

	correction := Correction{
		SignedAt: now,
	}
	errors := correction.Apply(lpa)

	assert.Equal(t, errors, []shared.FieldError{{Source: "/signedAt", Detail: "LPA Signed on date cannot be changed for online LPAs"}})
}

func TestCorrectionAttorneySignedAtChannel(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	lpa := &shared.Lpa{
		LpaInit: shared.LpaInit{
			Channel: "online",
			Attorneys: []shared.Attorney{
				{
					SignedAt: &yesterday,
				},
			},
		},
	}

	index := 0
	correction := Correction{
		Attorney: AttorneyCorrection{
			Index:    &index,
			SignedAt: now,
		},
	}
	errors := correction.Apply(lpa)

	assert.Equal(t, errors, []shared.FieldError{{Source: "/attorney/0/signedAt", Detail: "The attorney signed at date cannot be changed for online LPA"}})
}

func TestCorrectionApplyForCertificateProvider(t *testing.T) {
	twoDaysAgo := time.Now().Add(-48 * time.Hour)
	yesterday := time.Now().Add(-24 * time.Hour)

	lpa := &shared.Lpa{
		LpaInit: shared.LpaInit{
			CertificateProvider: shared.CertificateProvider{
				Person: shared.Person{
					FirstNames: "Branson",
					LastName:   "Conn",
				},
				Address: shared.Address{
					Line1:    "9 Kutch Meadows",
					Line2:    "Cummerata",
					Town:     "West Blick",
					Postcode: "YX97 3HZ",
					Country:  "UK",
				},
				Email:    "Branson.Conn@example.com",
				Phone:    "01977 67513",
				SignedAt: &twoDaysAgo,
			},
		},
	}

	certificateProviderCorrection := CertificateProviderCorrection{
		FirstNames: "Lynn",
		LastName:   "Christiansen",
		Address: shared.Address{
			Line1:    "653 Prosacco Avenue",
			Town:     "Long Larkin",
			Postcode: "RC18 6RZ",
			Country:  "UK",
		},
		Email:    "Lynn.Christiansen@example.com",
		Phone:    "01003 19993",
		SignedAt: yesterday,
	}

	correction := Correction{
		CertificateProvider: certificateProviderCorrection,
	}

	errors := correction.Apply(lpa)

	assert.Empty(t, errors)
	assert.Equal(t, correction.CertificateProvider.FirstNames, lpa.CertificateProvider.FirstNames)
	assert.Equal(t, correction.CertificateProvider.LastName, lpa.CertificateProvider.LastName)
	assert.Equal(t, correction.CertificateProvider.Address, lpa.CertificateProvider.Address)
	assert.Equal(t, correction.CertificateProvider.Email, lpa.CertificateProvider.Email)
	assert.Equal(t, correction.CertificateProvider.Phone, lpa.CertificateProvider.Phone)
	assert.Equal(t, correction.CertificateProvider.SignedAt, *lpa.CertificateProvider.SignedAt)
}

func TestCorrectionApplyForCertificateProviderSignedAtChannel(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	lpa := &shared.Lpa{
		LpaInit: shared.LpaInit{
			Channel: "online",
			CertificateProvider: shared.CertificateProvider{
				SignedAt: &yesterday,
			},
		},
	}

	correction := Correction{
		CertificateProvider: CertificateProviderCorrection{
			SignedAt: now,
		},
	}
	errors := correction.Apply(lpa)

	assert.Equal(t, errors, []shared.FieldError{{
		Source: "/certificateProvider/signedAt",
		Detail: "The Certificate Provider Signed on date cannot be changed for online LPAs",
	}})
}

func TestValidateCorrection(t *testing.T) {
	const fieldRequired = "field is required"
	now := time.Now().Round(time.Millisecond).UTC()

	testcases := map[string]struct {
		changes  []shared.Change
		lpa      *shared.Lpa
		expected Correction
		errors   []shared.FieldError
	}{
		"valid donor update": {
			changes: []shared.Change{
				{Key: "/donor/firstNames", New: json.RawMessage(`"Jane"`), Old: jsonNull},
				{Key: "/donor/lastName", New: json.RawMessage(`"Doe"`), Old: jsonNull},
				{Key: "/donor/otherNamesKnownBy", New: json.RawMessage(`"Janey"`), Old: jsonNull},
				{Key: "/donor/email", New: json.RawMessage(`"jane.doe@example.com"`), Old: jsonNull},
				{Key: "/donor/dateOfBirth", New: json.RawMessage(`"2000-01-01"`), Old: jsonNull},
				{Key: "/signedAt", New: json.RawMessage(`"` + now.Format(time.RFC3339Nano) + `"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Donor: shared.Donor{},
				},
			},
			expected: Correction{
				Donor: DonorCorrection{
					FirstNames:        "Jane",
					LastName:          "Doe",
					OtherNamesKnownBy: "Janey",
					Email:             "jane.doe@example.com",
					DateOfBirth:       createDate("2000-01-01"),
				},
				SignedAt: now,
			},
		},
		"valid attorney update": {
			changes: []shared.Change{
				{Key: "/attorneys/0/firstNames", New: json.RawMessage(`"Shanelle"`), Old: jsonNull},
				{Key: "/attorneys/0/lastName", New: json.RawMessage(`"Kerluke"`), Old: jsonNull},
				{Key: "/attorneys/0/dateOfBirth", New: json.RawMessage(`"1949-10-20"`), Old: jsonNull},
				{Key: "/attorneys/0/email", New: json.RawMessage(`"test@test.com"`), Old: jsonNull},
				{Key: "/attorneys/0/mobile", New: json.RawMessage(`"123456789"`), Old: jsonNull},
				{Key: "/attorneys/0/address/line1", New: json.RawMessage(`"13 Park Avenue"`), Old: jsonNull},
				{Key: "/attorneys/0/address/town", New: json.RawMessage(`"Clwyd"`), Old: jsonNull},
				{Key: "/attorneys/0/address/postcode", New: json.RawMessage(`"OH03 2LM"`), Old: jsonNull},
				{Key: "/attorneys/0/address/country", New: json.RawMessage(`"GB"`), Old: jsonNull},
				{Key: "/attorneys/0/signedAt", New: json.RawMessage(`"` + now.Format(time.RFC3339Nano) + `"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys: []shared.Attorney{
						{},
					},
				},
			},
			expected: Correction{
				Attorney: AttorneyCorrection{
					Index:       ptrTo(0),
					FirstNames:  "Shanelle",
					LastName:    "Kerluke",
					DateOfBirth: createDate("1949-10-20"),
					Email:       "test@test.com",
					Mobile:      "123456789",
					Address: shared.Address{
						Line1:    "13 Park Avenue",
						Town:     "Clwyd",
						Postcode: "OH03 2LM",
						Country:  "GB",
					},
					SignedAt: now,
				},
			},
		},
		"valid replacement attorney update": {
			changes: []shared.Change{
				{Key: "/attorneys/1/firstNames", New: json.RawMessage(`"Anthony"`), Old: jsonNull},
				{Key: "/attorneys/1/lastName", New: json.RawMessage(`"Leannon"`), Old: jsonNull},
				{Key: "/attorneys/1/dateOfBirth", New: json.RawMessage(`"1963-11-08"`), Old: jsonNull},
				{Key: "/attorneys/1/address/town", New: json.RawMessage(`"Cheshire"`), Old: jsonNull},
				{Key: "/attorneys/1/address/country", New: json.RawMessage(`"GB"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys: []shared.Attorney{
						{
							AppointmentType: shared.AppointmentTypeOriginal,
							Status:          shared.AttorneyStatusActive,
						},
						{
							AppointmentType: shared.AppointmentTypeReplacement,
							Status:          shared.AttorneyStatusInactive,
						},
					},
				},
			},
			expected: Correction{
				Attorney: AttorneyCorrection{
					Index:       ptrTo(1),
					FirstNames:  "Anthony",
					LastName:    "Leannon",
					DateOfBirth: createDate("1963-11-08"),
					Address: shared.Address{
						Town:    "Cheshire",
						Country: "GB",
					},
				},
			},
		},
		"valid certificate provider update": {
			changes: []shared.Change{
				{Key: "/certificateProvider/firstNames", New: json.RawMessage(`"Trinity"`), Old: jsonNull},
				{Key: "/certificateProvider/lastName", New: json.RawMessage(`"Monahan"`), Old: jsonNull},
				{Key: "/certificateProvider/email", New: json.RawMessage(`"Trinity.Monahan@example.com"`), Old: jsonNull},
				{Key: "/certificateProvider/phone", New: json.RawMessage(`"01697 233 415"`), Old: jsonNull},
				{Key: "/certificateProvider/signedAt", New: json.RawMessage(`"` + now.Format(time.RFC3339Nano) + `"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					CertificateProvider: shared.CertificateProvider{},
				},
			},
			expected: Correction{
				CertificateProvider: CertificateProviderCorrection{
					FirstNames: "Trinity",
					LastName:   "Monahan",
					Email:      "Trinity.Monahan@example.com",
					Phone:      "01697 233 415",
					SignedAt:   now,
				},
			},
		},
		"valid attorney decisions update": {
			changes: []shared.Change{
				{Key: "/howAttorneysMakeDecisions", New: json.RawMessage(`"jointly"`), Old: json.RawMessage(`"jointly-and-severally"`)},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys:                 []shared.Attorney{{Status: shared.AttorneyStatusActive}, {Status: shared.AttorneyStatusActive}},
					HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
				},
			},
			expected: Correction{
				HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointly,
			},
		},
		"valid replacement attorneys step in update": {
			changes: []shared.Change{
				{Key: "/howReplacementAttorneysStepIn", New: json.RawMessage(`"one-can-no-longer-act"`), Old: json.RawMessage(`"all-can-no-longer-act"`)},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys: []shared.Attorney{
						{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
					},
					HowAttorneysMakeDecisions:     shared.HowMakeDecisionsJointlyAndSeverally,
					HowReplacementAttorneysStepIn: shared.HowStepInAllCanNoLongerAct,
				},
			},
			expected: Correction{
				HowReplacementAttorneysStepIn: shared.HowStepInOneCanNoLongerAct,
			},
		},
		"valid replacement attorney decisions update": {
			changes: []shared.Change{
				{Key: "/howReplacementAttorneysMakeDecisions", New: json.RawMessage(`"jointly"`), Old: json.RawMessage(`"jointly-and-severally"`)},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys: []shared.Attorney{
						{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
						{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
					},
					HowReplacementAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
				},
			},
			expected: Correction{
				HowReplacementAttorneysMakeDecisions: shared.HowMakeDecisionsJointly,
			},
		},
		"valid replacement attorney decisions with details update": {
			changes: []shared.Change{
				{Key: "/howReplacementAttorneysMakeDecisions", New: json.RawMessage(`"jointly-for-some-severally-for-others"`), Old: json.RawMessage(`"jointly-and-severally"`)},
				{Key: "/howReplacementAttorneysMakeDecisionsDetails", New: json.RawMessage(`"blah"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys: []shared.Attorney{
						{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
						{Status: shared.AttorneyStatusInactive, AppointmentType: shared.AppointmentTypeReplacement},
					},
					HowReplacementAttorneysMakeDecisions: shared.HowMakeDecisionsJointlyAndSeverally,
				},
			},
			expected: Correction{
				HowReplacementAttorneysMakeDecisions:        shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
				HowReplacementAttorneysMakeDecisionsDetails: "blah",
			},
		},
		"valid life sustaining treatment update": {
			changes: []shared.Change{
				{Key: "/lifeSustainingTreatmentOption", New: json.RawMessage(`"option-b"`), Old: json.RawMessage(`"option-a"`)},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					LpaType:                       shared.LpaTypePersonalWelfare,
					LifeSustainingTreatmentOption: shared.LifeSustainingTreatmentOptionA,
				},
			},
			expected: Correction{
				LifeSustainingTreatmentOption: shared.LifeSustainingTreatmentOptionB,
			},
		},
		"valid when can use update": {
			changes: []shared.Change{
				{Key: "/whenTheLpaCanBeUsed", New: json.RawMessage(`"when-capacity-lost"`), Old: json.RawMessage(`"when-has-capacity"`)},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					LpaType:             shared.LpaTypePropertyAndAffairs,
					WhenTheLpaCanBeUsed: shared.CanUseWhenHasCapacity,
				},
			},
			expected: Correction{
				WhenTheLpaCanBeUsed: shared.CanUseWhenCapacityLost,
			},
		},
		"missing required fields": {
			changes: []shared.Change{
				{Key: "/donor/firstNames", New: jsonNull, Old: jsonNull},
				{Key: "/donor/lastName", New: jsonNull, Old: jsonNull},
				{Key: "/attorneys/0/firstNames", New: jsonNull, Old: jsonNull},
				{Key: "/attorneys/0/lastName", New: jsonNull, Old: jsonNull},
				{Key: "/certificateProvider/firstNames", New: jsonNull, Old: jsonNull},
				{Key: "/certificateProvider/lastName", New: jsonNull, Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Donor:               shared.Donor{},
					Attorneys:           []shared.Attorney{{}},
					CertificateProvider: shared.CertificateProvider{},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0/new", Detail: fieldRequired},
				{Source: "/changes/1/new", Detail: fieldRequired},
				{Source: "/changes/2/new", Detail: fieldRequired},
				{Source: "/changes/3/new", Detail: fieldRequired},
				{Source: "/changes/4/new", Detail: fieldRequired},
				{Source: "/changes/5/new", Detail: fieldRequired},
			},
		},
		"invalid country": {
			changes: []shared.Change{
				{Key: "/donor/address/country", New: json.RawMessage(`"United Kingdom"`), Old: jsonNull},
				{Key: "/attorneys/0/address/country", New: json.RawMessage(`"United Kingdom"`), Old: jsonNull},
				{Key: "/certificateProvider/address/country", New: json.RawMessage(`"United Kingdom"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Donor: shared.Donor{
						Address: shared.Address{},
					},
					Attorneys: []shared.Attorney{{}},
					CertificateProvider: shared.CertificateProvider{
						Address: shared.Address{},
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0/new", Detail: "must be a valid ISO-3166-1 country code"},
				{Source: "/changes/1/new", Detail: "must be a valid ISO-3166-1 country code"},
				{Source: "/changes/2/new", Detail: "must be a valid ISO-3166-1 country code"},
			},
		},
		"cannot change attorney decisions": {
			changes: []shared.Change{
				{Key: "/howAttorneysMakeDecisions", New: json.RawMessage(`"jointly"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys: []shared.Attorney{{Status: shared.AttorneyStatusActive}},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0", Detail: "unexpected change provided"},
			},
		},
		"cannot change replacement attorney decisions": {
			changes: []shared.Change{
				{Key: "/howReplacementAttorneysStepIn", New: json.RawMessage(`"one-can-no-longer-act"`), Old: jsonNull},
				{Key: "/howReplacementAttorneysMakeDecisions", New: json.RawMessage(`"jointly"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys: []shared.Attorney{{Status: shared.AttorneyStatusActive}},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0", Detail: "unexpected change provided"},
				{Source: "/changes/1", Detail: "unexpected change provided"},
			},
		},
		"cannot change life sustaining treatment": {
			changes: []shared.Change{
				{Key: "/lifeSustainingTreatmentOption", New: json.RawMessage(`"option-b"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					LpaType:                       shared.LpaTypePropertyAndAffairs,
					LifeSustainingTreatmentOption: shared.LifeSustainingTreatmentOptionA,
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0", Detail: "unexpected change provided"},
			},
		},
		"cannot change when can use": {
			changes: []shared.Change{
				{Key: "/whenTheLpaCanBeUsed", New: json.RawMessage(`"when-capacity-lost"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					LpaType:             shared.LpaTypePersonalWelfare,
					WhenTheLpaCanBeUsed: shared.CanUseWhenHasCapacity,
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0", Detail: "unexpected change provided"},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			correction, errors := validateCorrection(tc.changes, tc.lpa)
			assert.ElementsMatch(t, tc.errors, errors)
			if len(tc.errors) == 0 {
				assert.Equal(t, tc.expected, correction)
			}
		})
	}
}
