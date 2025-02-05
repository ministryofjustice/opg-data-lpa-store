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
		LPASignedAt: now,
	}

	errors := correction.Apply(lpa)

	assert.Empty(t, errors)
	assert.Equal(t, correction.Donor.FirstNames, lpa.Donor.FirstNames)
	assert.Equal(t, correction.Donor.LastName, lpa.Donor.LastName)
	assert.Equal(t, correction.Donor.OtherNamesKnownBy, lpa.Donor.OtherNamesKnownBy)
	assert.Equal(t, correction.Donor.DateOfBirth, lpa.Donor.DateOfBirth)
	assert.Equal(t, correction.Donor.Address, lpa.Donor.Address)
	assert.Equal(t, correction.Donor.Email, lpa.Donor.Email)
	assert.Equal(t, correction.LPASignedAt, lpa.SignedAt)
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
			Index:      &index,
			FirstNames: "Jane",
			LastName:   "Smith",
			Dob:        createDate("2000-11-11"),
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
	assert.Equal(t, correction.Attorney.Dob, lpa.Attorneys[index].DateOfBirth)
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
		LPASignedAt: now,
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

func TestValidateCorrection(t *testing.T) {
	now := time.Now()
	const fieldRequired = "field is required"

	testcases := map[string]struct {
		changes []shared.Change
		lpa     *shared.Lpa
		errors  []shared.FieldError
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
						shared.Attorney{},
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
							Status:          shared.AttorneyStatusReplacement,
						},
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
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Donor:     shared.Donor{},
					Attorneys: []shared.Attorney{{}},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0/new", Detail: fieldRequired},
				{Source: "/changes/1/new", Detail: fieldRequired},
				{Source: "/changes/2/new", Detail: fieldRequired},
				{Source: "/changes/3/new", Detail: fieldRequired},
			},
		},
		"invalid country": {
			changes: []shared.Change{
				{Key: "/donor/address/country", New: json.RawMessage(`"United Kingdom"`), Old: jsonNull},
				{Key: "/attorneys/0/address/country", New: json.RawMessage(`"United Kingdom"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Donor: shared.Donor{
						Address: shared.Address{},
					},
					Attorneys: []shared.Attorney{{}},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0/new", Detail: "must be a valid ISO-3166-1 country code"},
				{Source: "/changes/1/new", Detail: "must be a valid ISO-3166-1 country code"},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			_, errors := validateCorrection(tc.changes, tc.lpa)
			assert.ElementsMatch(t, tc.errors, errors)
		})
	}
}
