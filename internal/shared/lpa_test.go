package shared

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLpaInitMarshalJSON(t *testing.T) {
	expected := `{
"lpaType":"","channel":"",
"donor":{"uid":"","firstNames":"","lastName":"","address":{"line1":"","country":""},"dateOfBirth":"","email":"","contactLanguagePreference":""},
"attorneys":null,
"certificateProvider":{"uid":"","firstNames":"","lastName":"","address":{"line1":"","country":""},"email":"","phone":"","channel":""},
"signedAt":"0001-01-01T00:00:00Z"
}`

	data, _ := json.Marshal(LpaInit{})
	assert.JSONEq(t, expected, string(data))
}

func TestAttorneysGet(t *testing.T) {
	testCases := map[string]struct {
		attorneys        []Attorney
		expectedAttorney Attorney
		uid              string
		expectedFound    bool
	}{
		"found": {
			attorneys: []Attorney{
				{Person: Person{UID: "abc", FirstNames: "a"}},
				{Person: Person{UID: "xyz", FirstNames: "b"}},
			},
			expectedAttorney: Attorney{Person: Person{UID: "xyz", FirstNames: "b"}},
			uid:              "xyz",
			expectedFound:    true,
		},
		"not found": {
			attorneys: []Attorney{
				{Person: Person{UID: "abc", FirstNames: "a"}},
				{Person: Person{UID: "xyz", FirstNames: "b"}},
			},
			expectedAttorney: Attorney{},
			uid:              "not-a-match",
			expectedFound:    false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := &Lpa{LpaInit: LpaInit{Attorneys: tc.attorneys}}
			a, found := lpa.GetAttorney(tc.uid)

			assert.Equal(t, tc.expectedFound, found)
			assert.Equal(t, tc.expectedAttorney, a)
		})
	}
}

func TestAttorneysPut(t *testing.T) {
	testCases := map[string]struct {
		attorneys         []Attorney
		expectedAttorneys []Attorney
		updatedAttorney   Attorney
	}{
		"does not exist": {
			attorneys: []Attorney{
				{Person: Person{UID: "abc", FirstNames: "a"}},
			},
			expectedAttorneys: []Attorney{
				{Person: Person{UID: "abc", FirstNames: "a"}},
				{Person: Person{UID: "xyz", FirstNames: "b"}},
			},
			updatedAttorney: Attorney{Person: Person{UID: "xyz", FirstNames: "b"}},
		},
		"exists": {
			attorneys: []Attorney{
				{Person: Person{UID: "abc", FirstNames: "a"}},
				{Person: Person{UID: "xyz", FirstNames: "b"}},
			},
			expectedAttorneys: []Attorney{
				{Person: Person{UID: "abc", FirstNames: "a"}},
				{Person: Person{UID: "xyz", FirstNames: "z"}},
			},
			updatedAttorney: Attorney{Person: Person{UID: "xyz", FirstNames: "z"}},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := &Lpa{LpaInit: LpaInit{Attorneys: tc.attorneys}}
			lpa.PutAttorney(tc.updatedAttorney)

			assert.Equal(t, tc.expectedAttorneys, lpa.Attorneys)
		})
	}
}
