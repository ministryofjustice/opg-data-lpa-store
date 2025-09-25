package main

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createCustomDate(date string) shared.Date {
	d := shared.Date{}
	_ = d.UnmarshalText([]byte(date))
	return d
}

func pointerTo[T any](v T) *T {
	return &v
}

func TestPostRegistrationCorrectionApply(t *testing.T) {
	testcases := map[string]struct {
		correction PostRegistrationCorrection
		lpa        *shared.Lpa
		expected   *shared.Lpa
		errors     []shared.FieldError
	}{
		"decisions correction": {
			correction: PostRegistrationCorrection{
				AttorneyAppointmentType: AttorneyAppointmentPostRegistrationCorrection{
					shared.AttorneyAppointmentTypeCorrection{
						HowAttorneysMakeDecisions:                   shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
						HowAttorneysMakeDecisionsDetails:            "this way",
						HowReplacementAttorneysMakeDecisions:        shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
						HowReplacementAttorneysMakeDecisionsDetails: "that way",
						HowReplacementAttorneysStepIn:               shared.HowStepInAnotherWay,
						HowReplacementAttorneysStepInDetails:        "another way",
						WhenTheLpaCanBeUsed:                         shared.CanUseWhenCapacityLost,
						LifeSustainingTreatmentOption:               shared.LifeSustainingTreatmentOptionA,
					},
				},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions:                     shared.HowMakeDecisionsJointly,
					HowAttorneysMakeDecisionsIsDefault:            true,
					HowReplacementAttorneysMakeDecisions:          shared.HowMakeDecisionsJointly,
					HowReplacementAttorneysMakeDecisionsIsDefault: true,
					HowReplacementAttorneysStepIn:                 shared.HowStepInAllCanNoLongerAct,
					WhenTheLpaCanBeUsed:                           shared.CanUseWhenHasCapacity,
					WhenTheLpaCanBeUsedIsDefault:                  true,
					LifeSustainingTreatmentOption:                 shared.LifeSustainingTreatmentOptionB,
					LifeSustainingTreatmentOptionIsDefault:        true,
					Attorneys:                                     []shared.Attorney{},
				},
			},
			expected: &shared.Lpa{
				LpaInit: shared.LpaInit{
					HowAttorneysMakeDecisions:                   shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
					HowAttorneysMakeDecisionsDetails:            "this way",
					HowReplacementAttorneysMakeDecisions:        shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
					HowReplacementAttorneysMakeDecisionsDetails: "that way",
					HowReplacementAttorneysStepIn:               shared.HowStepInAnotherWay,
					HowReplacementAttorneysStepInDetails:        "another way",
					WhenTheLpaCanBeUsed:                         shared.CanUseWhenCapacityLost,
					LifeSustainingTreatmentOption:               shared.LifeSustainingTreatmentOptionA,
					CertificateProvider:                         shared.CertificateProvider{},
					Attorneys:                                   []shared.Attorney{},
				},
			},
		},
		"donor correction": {
			correction: PostRegistrationCorrection{
				Donor: DonorPostRegistrationCorrection{
					shared.DonorCorrection{
						FirstNames:        "Jane",
						LastName:          "Smith",
						OtherNamesKnownBy: "Janey",
						Address: shared.Address{
							Line1:    "456 Another St",
							Town:     "Othertown",
							Postcode: "B22 A11",
							Country:  "IE",
						},
						Email: "jane.smith@example.com",
					},
				},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Donor: shared.Donor{
						Person: shared.Person{
							FirstNames: "donor-firstname",
							LastName:   "donor-lastname",
						},
						OtherNamesKnownBy: "donor-otherNames",
						DateOfBirth:       createCustomDate("1990-01-02"),
						Address: shared.Address{
							Line1:    "123 Main St",
							Town:     "Anytown",
							Postcode: "A11 B22",
							Country:  "IE",
						},
						Email: "john.doe@example.com",
					},
					Attorneys: []shared.Attorney{},
				},
			},
			expected: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Donor: shared.Donor{
						Person: shared.Person{
							FirstNames: "Jane",
							LastName:   "Smith",
						},
						OtherNamesKnownBy: "Janey",
						DateOfBirth:       createCustomDate("1990-01-02"),
						Address: shared.Address{
							Line1:    "456 Another St",
							Town:     "Othertown",
							Postcode: "B22 A11",
							Country:  "IE",
						},
						Email: "jane.smith@example.com",
					},
					CertificateProvider: shared.CertificateProvider{},
					Attorneys:           []shared.Attorney{},
				},
			},
		},
		"certificate provider correction": {
			correction: PostRegistrationCorrection{
				CertificateProvider: CertificateProviderPostRegistrationCorrection{
					shared.CertificateProviderCorrection{
						FirstNames: "Lynn",
						LastName:   "Christiansen",
						Address: shared.Address{
							Line1:    "653 Prosacco Avenue",
							Town:     "Long Larkin",
							Postcode: "RC18 6RZ",
							Country:  "UK",
						},
						Email: "Lynn.Christiansen@example.com",
						Phone: "01003 19993",
					},
				},
			},
			lpa: &shared.Lpa{
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
						Email: "Branson.Conn@example.com",
						Phone: "01977 67513",
					},
				},
			},
			expected: &shared.Lpa{
				LpaInit: shared.LpaInit{
					CertificateProvider: shared.CertificateProvider{
						Person: shared.Person{
							FirstNames: "Lynn",
							LastName:   "Christiansen",
						},
						Address: shared.Address{
							Line1:    "653 Prosacco Avenue",
							Town:     "Long Larkin",
							Postcode: "RC18 6RZ",
							Country:  "UK",
						},
						Email: "Lynn.Christiansen@example.com",
						Phone: "01003 19993",
					},
				},
			},
		},
		"attorney correction": {
			correction: PostRegistrationCorrection{
				Attorney: AttorneyPostRegistrationCorrection{
					shared.AttorneyCorrection{
						Index:       pointerTo(1),
						FirstNames:  "Jane",
						LastName:    "Smith",
						DateOfBirth: createCustomDate("2000-11-11"),
						Address: shared.Address{
							Line1:    "456 Another St",
							Town:     "Othertown",
							Postcode: "B22 A11",
							Country:  "GB",
						},
						Email:  "jane.smith@example.com",
						Mobile: "987654321",
					},
				},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys: []shared.Attorney{
						{},
						{
							Person: shared.Person{
								FirstNames: "attorney-firstname",
								LastName:   "attorney-lastname",
							},
							DateOfBirth: createCustomDate("1990-01-02"),
							Address: shared.Address{
								Line1:    "123 Main St",
								Town:     "Anytown",
								Postcode: "A11 B22",
								Country:  "IE",
							},
							Email:  "test@test.com",
							Mobile: "0123456789",
						},
					},
				},
			},
			expected: &shared.Lpa{
				LpaInit: shared.LpaInit{
					CertificateProvider: shared.CertificateProvider{},
					Attorneys: []shared.Attorney{
						{},
						{
							Person: shared.Person{
								FirstNames: "Jane",
								LastName:   "Smith",
							},
							DateOfBirth: createCustomDate("2000-11-11"),
							Address: shared.Address{
								Line1:    "456 Another St",
								Town:     "Othertown",
								Postcode: "B22 A11",
								Country:  "GB",
							},
							Email:  "jane.smith@example.com",
							Mobile: "987654321",
						},
					},
				},
			},
		},
		"restrictions and conditions correction": {
			correction: PostRegistrationCorrection{
				RestrictionsAndConditions: "I live in Edinburgh",
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					RestrictionsAndConditions: "I always live in Edinburgh",
				},
			},
			expected: &shared.Lpa{
				LpaInit: shared.LpaInit{
					RestrictionsAndConditions: "I live in Edinburgh",
				},
			},
		},
	}

	for scenario, tc := range testcases {
		t.Run(scenario, func(t *testing.T) {
			errors := tc.correction.Apply(tc.lpa)

			if len(tc.errors) > 0 {
				assert.Equal(t, tc.errors, errors)
			} else {
				assert.Equal(t, tc.expected, tc.lpa)
			}
		})
	}
}

func TestValidatePostRegistrationCorrection(t *testing.T) {
	const fieldRequired = "field is required"

	testcases := map[string]struct {
		changes  []shared.Change
		lpa      *shared.Lpa
		expected PostRegistrationCorrection
		errors   []shared.FieldError
	}{
		"valid donor update": {
			changes: []shared.Change{
				{Key: "/donor/firstNames", New: json.RawMessage(`"Jane"`), Old: jsonNull},
				{Key: "/donor/lastName", New: json.RawMessage(`"Doe"`), Old: jsonNull},
				{Key: "/donor/otherNamesKnownBy", New: json.RawMessage(`"Janey"`), Old: jsonNull},
				{Key: "/donor/email", New: json.RawMessage(`"jane.doe@example.com"`), Old: jsonNull},
				{Key: "/donor/dateOfBirth", New: json.RawMessage(`"2000-01-01"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Donor: shared.Donor{},
				},
			},
			expected: PostRegistrationCorrection{
				Donor: DonorPostRegistrationCorrection{
					shared.DonorCorrection{
						FirstNames:        "Jane",
						LastName:          "Doe",
						OtherNamesKnownBy: "Janey",
						Email:             "jane.doe@example.com",
						DateOfBirth:       createCustomDate("2000-01-01"),
					},
				},
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
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys: []shared.Attorney{
						{},
					},
				},
			},
			expected: PostRegistrationCorrection{
				Attorney: AttorneyPostRegistrationCorrection{
					shared.AttorneyCorrection{
						Index:       pointerTo(0),
						FirstNames:  "Shanelle",
						LastName:    "Kerluke",
						DateOfBirth: createCustomDate("1949-10-20"),
						Email:       "test@test.com",
						Mobile:      "123456789",
						Address: shared.Address{
							Line1:    "13 Park Avenue",
							Town:     "Clwyd",
							Postcode: "OH03 2LM",
							Country:  "GB",
						},
					},
				},
			},
		},
		"valid attorney update with UID reference": {
			changes: []shared.Change{
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/firstNames", New: json.RawMessage(`"Shanelle"`), Old: jsonNull},
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/lastName", New: json.RawMessage(`"Kerluke"`), Old: jsonNull},
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/dateOfBirth", New: json.RawMessage(`"1949-10-20"`), Old: jsonNull},
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/email", New: json.RawMessage(`"test@test.com"`), Old: jsonNull},
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/mobile", New: json.RawMessage(`"123456789"`), Old: jsonNull},
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/address/line1", New: json.RawMessage(`"13 Park Avenue"`), Old: jsonNull},
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/address/town", New: json.RawMessage(`"Clwyd"`), Old: jsonNull},
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/address/postcode", New: json.RawMessage(`"OH03 2LM"`), Old: jsonNull},
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/address/country", New: json.RawMessage(`"GB"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"}},
					},
				},
			},
			expected: PostRegistrationCorrection{
				Attorney: AttorneyPostRegistrationCorrection{
					shared.AttorneyCorrection{
						Index:       pointerTo(0),
						FirstNames:  "Shanelle",
						LastName:    "Kerluke",
						DateOfBirth: createCustomDate("1949-10-20"),
						Email:       "test@test.com",
						Mobile:      "123456789",
						Address: shared.Address{
							Line1:    "13 Park Avenue",
							Town:     "Clwyd",
							Postcode: "OH03 2LM",
							Country:  "GB",
						},
					},
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
			expected: PostRegistrationCorrection{
				Attorney: AttorneyPostRegistrationCorrection{
					shared.AttorneyCorrection{
						Index:       pointerTo(1),
						FirstNames:  "Anthony",
						LastName:    "Leannon",
						DateOfBirth: createCustomDate("1963-11-08"),
						Address: shared.Address{
							Town:    "Cheshire",
							Country: "GB",
						},
					},
				},
			},
		},
		"valid replacement attorney update with UID reference": {
			changes: []shared.Change{
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/firstNames", New: json.RawMessage(`"Anthony"`), Old: jsonNull},
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/lastName", New: json.RawMessage(`"Leannon"`), Old: jsonNull},
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/dateOfBirth", New: json.RawMessage(`"1963-11-08"`), Old: jsonNull},
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/address/town", New: json.RawMessage(`"Cheshire"`), Old: jsonNull},
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/address/country", New: json.RawMessage(`"GB"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys: []shared.Attorney{
						{
							AppointmentType: shared.AppointmentTypeOriginal,
							Status:          shared.AttorneyStatusActive,
						},
						{
							Person:          shared.Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"},
							AppointmentType: shared.AppointmentTypeReplacement,
							Status:          shared.AttorneyStatusInactive,
						},
					},
				},
			},
			expected: PostRegistrationCorrection{
				Attorney: AttorneyPostRegistrationCorrection{
					shared.AttorneyCorrection{
						Index:       pointerTo(1),
						FirstNames:  "Anthony",
						LastName:    "Leannon",
						DateOfBirth: createCustomDate("1963-11-08"),
						Address: shared.Address{
							Town:    "Cheshire",
							Country: "GB",
						},
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
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					CertificateProvider: shared.CertificateProvider{},
				},
			},
			expected: PostRegistrationCorrection{
				CertificateProvider: CertificateProviderPostRegistrationCorrection{
					shared.CertificateProviderCorrection{
						FirstNames: "Trinity",
						LastName:   "Monahan",
						Email:      "Trinity.Monahan@example.com",
						Phone:      "01697 233 415",
					},
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
			expected: PostRegistrationCorrection{
				AttorneyAppointmentType: AttorneyAppointmentPostRegistrationCorrection{
					shared.AttorneyAppointmentTypeCorrection{
						HowAttorneysMakeDecisions: shared.HowMakeDecisionsJointly,
					},
				},
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
			expected: PostRegistrationCorrection{
				AttorneyAppointmentType: AttorneyAppointmentPostRegistrationCorrection{
					shared.AttorneyAppointmentTypeCorrection{
						HowReplacementAttorneysStepIn: shared.HowStepInOneCanNoLongerAct,
					},
				},
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
			expected: PostRegistrationCorrection{
				AttorneyAppointmentType: AttorneyAppointmentPostRegistrationCorrection{
					shared.AttorneyAppointmentTypeCorrection{
						HowReplacementAttorneysMakeDecisions: shared.HowMakeDecisionsJointly,
					},
				},
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
			expected: PostRegistrationCorrection{
				AttorneyAppointmentType: AttorneyAppointmentPostRegistrationCorrection{
					shared.AttorneyAppointmentTypeCorrection{
						HowReplacementAttorneysMakeDecisions:        shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers,
						HowReplacementAttorneysMakeDecisionsDetails: "blah",
					},
				},
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
			expected: PostRegistrationCorrection{
				AttorneyAppointmentType: AttorneyAppointmentPostRegistrationCorrection{
					shared.AttorneyAppointmentTypeCorrection{
						LifeSustainingTreatmentOption: shared.LifeSustainingTreatmentOptionB,
					},
				},
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
			expected: PostRegistrationCorrection{
				AttorneyAppointmentType: AttorneyAppointmentPostRegistrationCorrection{
					shared.AttorneyAppointmentTypeCorrection{
						WhenTheLpaCanBeUsed: shared.CanUseWhenCapacityLost,
					},
				},
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
		"missing required fields with UID reference": {
			changes: []shared.Change{
				{Key: "/donor/firstNames", New: jsonNull, Old: jsonNull},
				{Key: "/donor/lastName", New: jsonNull, Old: jsonNull},
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/firstNames", New: jsonNull, Old: jsonNull},
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/lastName", New: jsonNull, Old: jsonNull},
				{Key: "/certificateProvider/firstNames", New: jsonNull, Old: jsonNull},
				{Key: "/certificateProvider/lastName", New: jsonNull, Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Donor:               shared.Donor{},
					Attorneys:           []shared.Attorney{{Person: shared.Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"}}},
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
		"invalid country with UID reference": {
			changes: []shared.Change{
				{Key: "/donor/address/country", New: json.RawMessage(`"United Kingdom"`), Old: jsonNull},
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/address/country", New: json.RawMessage(`"United Kingdom"`), Old: jsonNull},
				{Key: "/certificateProvider/address/country", New: json.RawMessage(`"United Kingdom"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Donor: shared.Donor{
						Address: shared.Address{},
					},
					Attorneys: []shared.Attorney{{Person: shared.Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"}}},
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
		"cannot change replacement attorneys step in": {
			changes: []shared.Change{
				{Key: "/howReplacementAttorneysStepIn", New: json.RawMessage(`"one-can-no-longer-act"`), Old: json.RawMessage(`"all-can-no-longer-act"`)},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys: []shared.Attorney{
						{Status: shared.AttorneyStatusRemoved, AppointmentType: shared.AppointmentTypeReplacement},
					},
					HowAttorneysMakeDecisions:     shared.HowMakeDecisionsJointlyAndSeverally,
					HowReplacementAttorneysStepIn: shared.HowStepInAllCanNoLongerAct,
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
			correction, errors := validatePostRegistrationCorrection(tc.changes, tc.lpa)
			assert.ElementsMatch(t, tc.errors, errors)
			if len(tc.errors) == 0 {
				assert.Equal(t, tc.expected, correction)
			}
		})
	}
}
