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

func TestCorrectionApply(t *testing.T) {
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
		DonorFirstNames: "Jane",
		DonorLastName:   "Smith",
		DonorOtherNames: "Janey",
		DonorDob:        createDate("2000-11-11"),
		DonorAddress: shared.Address{
			Line1:    "456 Another St",
			Town:     "Othertown",
			Postcode: "B22 A11",
			Country:  "IE",
		},
		DonorEmail:  "jane.smith@example.com",
		LPASignedAt: now,
	}

	errors := correction.Apply(lpa)

	assert.Empty(t, errors)
	assert.Equal(t, correction.DonorFirstNames, lpa.Donor.FirstNames)
	assert.Equal(t, correction.DonorLastName, lpa.Donor.LastName)
	assert.Equal(t, correction.DonorOtherNames, lpa.Donor.OtherNamesKnownBy)
	assert.Equal(t, correction.DonorDob, lpa.Donor.DateOfBirth)
	assert.Equal(t, correction.DonorAddress, lpa.Donor.Address)
	assert.Equal(t, correction.DonorEmail, lpa.Donor.Email)
	assert.Equal(t, correction.LPASignedAt, lpa.SignedAt)
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
		DonorFirstNames: "Jane",
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

func TestValidateCorrection(t *testing.T) {
	now := time.Now()

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
				{Key: "/attorneys/0/firstNames", New: json.RawMessage(`"Jane"`), Old: jsonNull},
				{Key: "/attorneys/0/lastName", New: json.RawMessage(`"Doe"`), Old: jsonNull},
				{Key: "/attorneys/0/dateOfBirth", New: json.RawMessage(`"2000-01-01"`), Old: jsonNull},
				{Key: "/attorneys/0/email", New: json.RawMessage(`"test@test.com"`), Old: jsonNull},
				{Key: "/attorneys/0/mobile", New: json.RawMessage(`"123456789"`), Old: jsonNull},
				{Key: "/attorneys/0/address/line1", New: json.RawMessage(`"123 Main St"`), Old: jsonNull},
				{Key: "/attorneys/0/address/town", New: json.RawMessage(`"City"`), Old: jsonNull},
				{Key: "/attorneys/0/address/country", New: json.RawMessage(`"GB"`), Old: jsonNull},
				{Key: "/attorneys/0/signedAt", New: json.RawMessage(`"` + now.Format(time.RFC3339Nano) + `"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Donor: shared.Donor{},
					Attorneys: []shared.Attorney{
						shared.Attorney{},
					},
				},
			},
		},
		"missing required fields": {
			changes: []shared.Change{
				{Key: "/donor/firstNames", New: jsonNull, Old: jsonNull},
				{Key: "/donor/lastName", New: jsonNull, Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Donor: shared.Donor{},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0/new", Detail: "field is required"},
				{Source: "/changes/1/new", Detail: "field is required"},
			},
		},
		"invalid country": {
			changes: []shared.Change{
				{Key: "/donor/address/country", New: json.RawMessage(`"United Kingdom"`), Old: jsonNull},
			},
			lpa: &shared.Lpa{
				LpaInit: shared.LpaInit{
					Donor: shared.Donor{
						Address: shared.Address{},
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0/new", Detail: "must be a valid ISO-3166-1 country code"},
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
