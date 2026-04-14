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

func TestCorrectionApply(t *testing.T) {
	now := time.Now()
	nowFormatted := time.Now().Format(time.RFC3339)
	yesterday := now.AddDate(0, 0, -1)
	twoDaysAgo := time.Now().AddDate(0, 0, -2)

	testcases := map[string]struct {
		correction Correction
		lpa        *shared.Lpa
		expected   *shared.Lpa
		errors     []shared.FieldError
	}{
		"registered cannot change anything": {
			correction: Correction{
				Donor: DonorPreRegistrationCorrection{
					shared.DonorCorrection{
						FirstNames: "Jane",
					},
				},
			},
			lpa: &shared.Lpa{
				Status: shared.LpaStatusRegistered,
				LpaInit: shared.LpaInit{
					Channel: "paper",
					Donor: shared.Donor{
						Person: shared.Person{
							FirstNames: "donor-firstname",
						},
					},
				},
			},
			errors: []shared.FieldError{{Source: "/type", Detail: "Cannot make corrections to a Registered LPA"}},
		},
		"online cannot change signed at": {
			correction: Correction{
				SignedAt: now,
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Channel:  "online",
					SignedAt: yesterday,
				},
			},
			errors: []shared.FieldError{{Source: "/signedAt", Detail: "LPA Signed on date cannot be changed for online LPAs"}},
		},
		"decisions correction": {
			correction: Correction{
				AttorneyAppointmentType: AttorneyAppointmentPreRegistrationCorrection{
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
					Attorneys:                                   []shared.Attorney{},
				},
			},
		},
		"donor correction": {
			correction: Correction{
				Donor: DonorPreRegistrationCorrection{
					shared.DonorCorrection{
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
				},
				SignedAt: now,
			},
			lpa: &shared.Lpa{
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
			},
			expected: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Donor: shared.Donor{
						Person: shared.Person{
							FirstNames: "Jane",
							LastName:   "Smith",
						},
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
					SignedAt:  now,
					Attorneys: []shared.Attorney{},
				},
				Notes: []shared.Note{
					{
						Type:     "DONOR_NAME_CHANGE_V1",
						Datetime: nowFormatted,
						Values: map[string]string{
							"newName": "Jane Smith",
						},
					},
					{
						Type:     "DONOR_DOB_CHANGE_V1",
						Datetime: nowFormatted,
						Values: map[string]string{
							"newDob": "2000-11-11",
						},
					},
				},
			},
		},
		"donor date of birth cannot change after identity check": {
			correction: Correction{
				Donor: DonorPreRegistrationCorrection{
					shared.DonorCorrection{
						DateOfBirth: createDate("2000-11-11"),
					},
				},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Channel: "online",
					Donor: shared.Donor{
						DateOfBirth: createDate("2002-12-12"),
						IdentityCheck: &shared.IdentityCheck{
							CheckedAt: yesterday,
							Type:      shared.IdentityCheckTypeOneLogin,
						},
					},
				},
			},
			errors: []shared.FieldError{{Source: "/donor/dateOfBirth", Detail: "The donor's date of birth cannot be changed once the identity check is complete"}},
		},
		"certificate provider correction": {
			correction: Correction{
				CertificateProvider: CertificateProviderPreRegistrationCorrection{
					shared.CertificateProviderCorrection{
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
						Email:    "Branson.Conn@example.com",
						Phone:    "01977 67513",
						SignedAt: &twoDaysAgo,
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
						Email:    "Lynn.Christiansen@example.com",
						Phone:    "01003 19993",
						SignedAt: &yesterday,
					},
				},
				Notes: []shared.Note{{
					Type:     "CERTIFICATE_PROVIDER_NAME_CHANGE_V1",
					Datetime: nowFormatted,
					Values: map[string]string{
						"newName": "Lynn Christiansen",
					},
				}},
			},
		},
		"certificate provider cannot change signed at": {
			correction: Correction{
				CertificateProvider: CertificateProviderPreRegistrationCorrection{
					shared.CertificateProviderCorrection{
						SignedAt: now,
					},
				},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Channel: "online",
					CertificateProvider: shared.CertificateProvider{
						SignedAt: &yesterday,
					},
				},
			},
			errors: []shared.FieldError{{
				Source: "/certificateProvider/signedAt",
				Detail: "The Certificate Provider Signed on date cannot be changed for online LPAs",
			}},
		},
		"attorney correction": {
			correction: Correction{
				Attorney: AttorneyPreRegistrationCorrection{
					shared.AttorneyCorrection{
						Index:       ptrTo(1),
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
			},
			expected: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys: []shared.Attorney{
						{},
						{
							Person: shared.Person{
								FirstNames: "Jane",
								LastName:   "Smith",
							},
							DateOfBirth: createDate("2000-11-11"),
							Address: shared.Address{
								Line1:    "456 Another St",
								Town:     "Othertown",
								Postcode: "B22 A11",
								Country:  "GB",
							},
							Email:    "jane.smith@example.com",
							Mobile:   "987654321",
							SignedAt: &now,
						},
					},
				},
			},
		},
		"trust corporation correction": {
			correction: Correction{
				TrustCorporation: TrustCorporationPreRegistrationCorrection{
					shared.TrustCorporationCorrection{
						Index:         ptrTo(0),
						Name:          "Webster Mraz Limited",
						CompanyNumber: "315724446",
						Email:         "Webster.Mraz@example.com",
						Address: shared.Address{
							Line1:    "518 Gussie Meadows",
							Town:     "Hamill",
							Postcode: "RD3 8OI",
							Country:  "GB",
						},
						Mobile: "0809 694 3813",
						Signatories: []shared.Signatory{{
							FirstNames:        "Raphael",
							LastName:          "Hansen",
							ProfessionalTitle: "Prof.",
							SignedAt:          now,
						}, {
							FirstNames:        "Marianna",
							LastName:          "Quigley",
							ProfessionalTitle: "Dr.",
							SignedAt:          now,
						}},
					},
				},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					TrustCorporations: []shared.TrustCorporation{{
						Name:          "Stacey Wilderman Limited",
						CompanyNumber: "096912473",
						Email:         "Stacey.Wilderman@example.com",
						Address: shared.Address{
							Line1:    "127 Vicarage Close",
							Town:     "Wehnerfield",
							Postcode: "WB74 0BF",
							Country:  "GB",
						},
						Status: shared.AttorneyStatusActive,
						Mobile: "0191 996 6889",
						Signatories: []shared.Signatory{{
							FirstNames:        "Miracle",
							LastName:          "Morar",
							ProfessionalTitle: "Dr.",
							SignedAt:          yesterday,
						}, {
							FirstNames:        "Miracle",
							LastName:          "Morar",
							ProfessionalTitle: "Dr.",
							SignedAt:          yesterday,
						}},
					}},
				},
			},
			expected: &shared.Lpa{
				LpaInit: shared.LpaInit{
					TrustCorporations: []shared.TrustCorporation{{
						Name:          "Webster Mraz Limited",
						CompanyNumber: "315724446",
						Email:         "Webster.Mraz@example.com",
						Address: shared.Address{
							Line1:    "518 Gussie Meadows",
							Town:     "Hamill",
							Postcode: "RD3 8OI",
							Country:  "GB",
						},
						Status: shared.AttorneyStatusActive,
						Mobile: "0809 694 3813",
						Signatories: []shared.Signatory{{
							FirstNames:        "Raphael",
							LastName:          "Hansen",
							ProfessionalTitle: "Prof.",
							SignedAt:          yesterday,
						}, {
							FirstNames:        "Marianna",
							LastName:          "Quigley",
							ProfessionalTitle: "Dr.",
							SignedAt:          yesterday,
						}},
					}},
				},
			},
		},
		"attorney cannot change signed at": {
			correction: Correction{
				Attorney: AttorneyPreRegistrationCorrection{
					shared.AttorneyCorrection{
						Index:    ptrTo(0),
						SignedAt: now,
					},
				},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Channel: "online",
					Attorneys: []shared.Attorney{
						{
							SignedAt: &yesterday,
						},
					},
				},
			},
			errors: []shared.FieldError{{Source: "/attorney/0/signedAt", Detail: "The attorney signed at date cannot be changed for online LPA"}},
		},
		"authorised signatory": {
			correction: Correction{
				AuthorisedSignatory: AuthorisedSignatoryPreRegistrationCorrection{
					shared.AuthorisedSignatoryCorrection{
						FirstNames: "Jamar",
						LastName:   "Dakota",
					},
				},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					AuthorisedSignatory: &shared.AuthorisedSignatory{
						Person: shared.Person{
							FirstNames: "Mafalda",
							LastName:   "Kuhic",
						},
					},
				},
			},
			expected: &shared.Lpa{
				LpaInit: shared.LpaInit{
					AuthorisedSignatory: &shared.AuthorisedSignatory{
						Person: shared.Person{
							FirstNames: "Jamar",
							LastName:   "Dakota",
						},
					},
				},
			},
		},
		"independent witness": {
			correction: Correction{
				IndependentWitness: IndependentWitnessPreRegistrationCorrection{
					shared.IndependentWitnessCorrection{
						FirstNames: "Shaniya",
						LastName:   "Rowan",
						Phone:      "0955 305 0174",
						Address: shared.Address{
							Line1:    "798 Genevieve Drove",
							Line2:    "Hudsonwood",
							Line3:    "Hellerwick",
							Town:     "Wuckert",
							Postcode: "AN85 0JJ",
							Country:  "GB",
						},
					},
				},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					IndependentWitness: &shared.IndependentWitness{
						Person: shared.Person{
							FirstNames: "Donald",
							LastName:   "Rowe",
						},
						Phone: "0115 592 8043",
						Address: shared.Address{
							Line1:    "3 Station Road",
							Line2:    "Ziemann",
							Line3:    "Wuckert",
							Town:     "Avon",
							Postcode: "NE0 8KM",
							Country:  "GB",
						},
					},
				},
			},
			expected: &shared.Lpa{
				LpaInit: shared.LpaInit{
					IndependentWitness: &shared.IndependentWitness{
						Person: shared.Person{
							FirstNames: "Shaniya",
							LastName:   "Rowan",
						},
						Phone: "0955 305 0174",
						Address: shared.Address{
							Line1:    "798 Genevieve Drove",
							Line2:    "Hudsonwood",
							Line3:    "Hellerwick",
							Town:     "Wuckert",
							Postcode: "AN85 0JJ",
							Country:  "GB",
						},
					},
				},
			},
		},
		"witnessed by": {
			correction: Correction{
				WitnessedBy: WitnessedByPreRegistrationCorrection{
					shared.WitnessedByCorrection{
						WitnessedByCertificateProviderAt: now,
						WitnessedByIndependentWitnessAt:  now,
					},
				},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					WitnessedByCertificateProviderAt: yesterday,
					WitnessedByIndependentWitnessAt:  &yesterday,
				},
			},
			expected: &shared.Lpa{
				LpaInit: shared.LpaInit{
					WitnessedByCertificateProviderAt: now,
					WitnessedByIndependentWitnessAt:  &now,
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

func TestValidateCorrection(t *testing.T) {
	const fieldRequired = "field is required"
	now := time.Now().Round(time.Millisecond).UTC()

	testcases := map[string]struct {
		changes  []shared.Change
		lpa      *shared.Lpa
		expected Correction
		errors   []shared.FieldError
	}{
		"no changes provided": {
			lpa:    &shared.Lpa{},
			errors: []shared.FieldError{{Source: "/changes", Detail: "no changes provided"}},
		},
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
				Donor: DonorPreRegistrationCorrection{
					shared.DonorCorrection{
						FirstNames:        "Jane",
						LastName:          "Doe",
						OtherNamesKnownBy: "Janey",
						Email:             "jane.doe@example.com",
						DateOfBirth:       createDate("2000-01-01"),
					},
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
				Attorney: AttorneyPreRegistrationCorrection{
					shared.AttorneyCorrection{
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
				{Key: "/attorneys/9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d/signedAt", New: json.RawMessage(`"` + now.Format(time.RFC3339Nano) + `"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys: []shared.Attorney{
						{Person: shared.Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"}},
					},
				},
			},
			expected: Correction{
				Attorney: AttorneyPreRegistrationCorrection{
					shared.AttorneyCorrection{
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
				Attorney: AttorneyPreRegistrationCorrection{
					shared.AttorneyCorrection{
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
			expected: Correction{
				Attorney: AttorneyPreRegistrationCorrection{
					shared.AttorneyCorrection{
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
				CertificateProvider: CertificateProviderPreRegistrationCorrection{
					shared.CertificateProviderCorrection{
						FirstNames: "Trinity",
						LastName:   "Monahan",
						Email:      "Trinity.Monahan@example.com",
						Phone:      "01697 233 415",
						SignedAt:   now,
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
			expected: Correction{
				AttorneyAppointmentType: AttorneyAppointmentPreRegistrationCorrection{
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
			expected: Correction{
				AttorneyAppointmentType: AttorneyAppointmentPreRegistrationCorrection{
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
			expected: Correction{
				AttorneyAppointmentType: AttorneyAppointmentPreRegistrationCorrection{
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
			expected: Correction{
				AttorneyAppointmentType: AttorneyAppointmentPreRegistrationCorrection{
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
			expected: Correction{
				AttorneyAppointmentType: AttorneyAppointmentPreRegistrationCorrection{
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
			expected: Correction{
				AttorneyAppointmentType: AttorneyAppointmentPreRegistrationCorrection{
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
				{Key: "/trustCorporations/0/name", New: jsonNull, Old: jsonNull},
				{Key: "/trustCorporations/0/companyNumber", New: jsonNull, Old: jsonNull},
				{Key: "/authorisedSignatory/firstNames", New: jsonNull, Old: jsonNull},
				{Key: "/authorisedSignatory/lastName", New: jsonNull, Old: jsonNull},
				{Key: "/independentWitness/firstNames", New: jsonNull, Old: jsonNull},
				{Key: "/independentWitness/lastName", New: jsonNull, Old: jsonNull},
				{Key: "/independentWitness/phone", New: jsonNull, Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Donor:               shared.Donor{},
					Attorneys:           []shared.Attorney{{}},
					CertificateProvider: shared.CertificateProvider{},
					TrustCorporations:   []shared.TrustCorporation{{}},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0/new", Detail: fieldRequired},
				{Source: "/changes/1/new", Detail: fieldRequired},
				{Source: "/changes/2/new", Detail: fieldRequired},
				{Source: "/changes/3/new", Detail: fieldRequired},
				{Source: "/changes/4/new", Detail: fieldRequired},
				{Source: "/changes/5/new", Detail: fieldRequired},
				{Source: "/changes/6/new", Detail: fieldRequired},
				{Source: "/changes/7/new", Detail: fieldRequired},
				{Source: "/changes/8/new", Detail: fieldRequired},
				{Source: "/changes/9/new", Detail: fieldRequired},
				{Source: "/changes/10/new", Detail: fieldRequired},
				{Source: "/changes/11/new", Detail: fieldRequired},
				{Source: "/changes/12/new", Detail: fieldRequired},
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
				{Key: "/trustCorporations/0/address/country", New: json.RawMessage(`"United Kingdom"`), Old: jsonNull},
				{Key: "/independentWitness/address/country", New: json.RawMessage(`"United Kingdom"`), Old: jsonNull},
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
					TrustCorporations: []shared.TrustCorporation{{}},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0/new", Detail: "must be a valid ISO-3166-1 country code"},
				{Source: "/changes/1/new", Detail: "must be a valid ISO-3166-1 country code"},
				{Source: "/changes/2/new", Detail: "must be a valid ISO-3166-1 country code"},
				{Source: "/changes/3/new", Detail: "must be a valid ISO-3166-1 country code"},
				{Source: "/changes/4/new", Detail: "must be a valid ISO-3166-1 country code"},
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
		"valid trust corporation update": {
			changes: []shared.Change{
				{Key: "/trustCorporations/0/name", New: json.RawMessage(`"Burley Gottlieb Limited"`), Old: json.RawMessage(`"Cecil Harper Limited"`)},
				{Key: "/trustCorporations/0/companyNumber", New: json.RawMessage(`"710228347"`), Old: json.RawMessage(`"634843055"`)},
				{Key: "/trustCorporations/0/address/line1", New: json.RawMessage(`"7 Paxton Drove"`), Old: json.RawMessage(`"37 Niko Wynd"`)},
				{Key: "/trustCorporations/0/address/line2", New: json.RawMessage(`"Kessler"`), Old: json.RawMessage(`"Upper Vonford"`)},
				{Key: "/trustCorporations/0/address/line3", New: json.RawMessage(`"Nienowbury"`), Old: json.RawMessage(`"Upton Stanton"`)},
				{Key: "/trustCorporations/0/address/town", New: json.RawMessage(`"Cormier"`), Old: json.RawMessage(`"Conroy"`)},
				{Key: "/trustCorporations/0/address/postcode", New: json.RawMessage(`"DB2 2RT"`), Old: json.RawMessage(`"SE64 1ZD"`)},
				{Key: "/trustCorporations/0/address/country", New: json.RawMessage(`"GB"`), Old: json.RawMessage(`"GB"`)},
				{Key: "/trustCorporations/0/signatories/0/firstNames", New: json.RawMessage(`"Charlie"`), Old: json.RawMessage(`"Virginie"`)},
				{Key: "/trustCorporations/0/signatories/0/lastName", New: json.RawMessage(`"Pollich"`), Old: json.RawMessage(`"Reichert"`)},
				{Key: "/trustCorporations/0/signatories/1/firstNames", New: json.RawMessage(`"Norene"`), Old: json.RawMessage(`"Modesto"`)},
				{Key: "/trustCorporations/0/signatories/1/lastName", New: json.RawMessage(`"Marks"`), Old: json.RawMessage(`"Adams"`)},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					TrustCorporations: []shared.TrustCorporation{{
						Name:          "Cecil Harper Limited",
						CompanyNumber: "634843055",
						Address: shared.Address{
							Line1:    "37 Niko Wynd",
							Line2:    "Upper Vonford",
							Line3:    "Upton Stanton",
							Town:     "Conroy",
							Postcode: "SE64 1ZD",
							Country:  "GB",
						},
						Signatories: []shared.Signatory{{
							FirstNames: "Virginie",
							LastName:   "Reichert",
						}, {
							FirstNames: "Modesto",
							LastName:   "Adams",
						}},
					}},
				},
			},
			expected: Correction{
				TrustCorporation: TrustCorporationPreRegistrationCorrection{
					shared.TrustCorporationCorrection{
						Index:         ptrTo(0),
						Name:          "Burley Gottlieb Limited",
						CompanyNumber: "710228347",
						Address: shared.Address{
							Line1:    "7 Paxton Drove",
							Line2:    "Kessler",
							Line3:    "Nienowbury",
							Town:     "Cormier",
							Postcode: "DB2 2RT",
							Country:  "GB",
						},
						Signatories: []shared.Signatory{
							{
								FirstNames: "Charlie",
								LastName:   "Pollich",
							},
							{
								FirstNames: "Norene",
								LastName:   "Marks",
							},
						},
					},
				},
			},
		},
		"valid authorised signatory update": {
			changes: []shared.Change{
				{Key: "/authorisedSignatory/firstNames", New: json.RawMessage(`"Orlando"`), Old: jsonNull},
				{Key: "/authorisedSignatory/lastName", New: json.RawMessage(`"Breitenberg"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{},
			},
			expected: Correction{
				AuthorisedSignatory: AuthorisedSignatoryPreRegistrationCorrection{
					AuthorisedSignatoryCorrection: shared.AuthorisedSignatoryCorrection{
						FirstNames: "Orlando",
						LastName:   "Breitenberg",
					},
				},
			},
		},
		"valid independent witness update": {
			changes: []shared.Change{
				{Key: "/independentWitness/firstNames", New: json.RawMessage(`"Cleora"`), Old: jsonNull},
				{Key: "/independentWitness/lastName", New: json.RawMessage(`"Koss"`), Old: jsonNull},
				{Key: "/independentWitness/phone", New: json.RawMessage(`"016977 8334"`), Old: jsonNull},
				{Key: "/independentWitness/address/line1", New: json.RawMessage(`"48 Upton Mead"`), Old: jsonNull},
				{Key: "/independentWitness/address/line2", New: json.RawMessage(`"Willms"`), Old: jsonNull},
				{Key: "/independentWitness/address/line3", New: json.RawMessage(`"Kertzmannstone"`), Old: jsonNull},
				{Key: "/independentWitness/address/town", New: json.RawMessage(`"Devon"`), Old: jsonNull},
				{Key: "/independentWitness/address/postcode", New: json.RawMessage(`"BA7 5IB"`), Old: jsonNull},
				{Key: "/independentWitness/address/country", New: json.RawMessage(`"GB"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{},
			},
			expected: Correction{
				IndependentWitness: IndependentWitnessPreRegistrationCorrection{
					IndependentWitnessCorrection: shared.IndependentWitnessCorrection{
						FirstNames: "Cleora",
						LastName:   "Koss",
						Phone:      "016977 8334",
						Address: shared.Address{
							Line1:    "48 Upton Mead",
							Line2:    "Willms",
							Line3:    "Kertzmannstone",
							Town:     "Devon",
							Postcode: "BA7 5IB",
							Country:  "GB",
						},
					},
				},
			},
		},
		"invalid date values": {
			changes: []shared.Change{
				{Key: "/witnessedByCertificateProviderAt", New: json.RawMessage(`"Invalid"`), Old: jsonNull},
				{Key: "/witnessedByIndependentWitnessAt", New: json.RawMessage(`"Invalid"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0/new", Detail: "unexpected type"},
				{Source: "/changes/1/new", Detail: "unexpected type"},
			},
		},
		"valid witnessed by update": {
			changes: []shared.Change{
				{Key: "/witnessedByCertificateProviderAt", New: json.RawMessage(`"` + now.Format(time.RFC3339Nano) + `"`), Old: jsonNull},
				{Key: "/witnessedByIndependentWitnessAt", New: json.RawMessage(`"` + now.Format(time.RFC3339Nano) + `"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{},
			},
			expected: Correction{
				WitnessedBy: WitnessedByPreRegistrationCorrection{
					shared.WitnessedByCorrection{
						WitnessedByCertificateProviderAt: now,
						WitnessedByIndependentWitnessAt:  now,
					},
				},
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
