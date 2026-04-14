package shared

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLpaInitMarshalJSON(t *testing.T) {
	expected := `{
"lpaType":"","channel":"","language":"",
"donor":{"uid":"","firstNames":"","lastName":"","address":{"line1":"","country":""},"dateOfBirth":"","contactLanguagePreference":""},
"attorneys":null,
"certificateProvider":{"uid":"","firstNames":"","lastName":"","address":{"line1":"","country":""},"phone":"","channel":""},
"signedAt":"0001-01-01T00:00:00Z","witnessedByCertificateProviderAt":"0001-01-01T00:00:00Z"
}`

	data, _ := json.Marshal(LpaInit{})
	assert.JSONEq(t, expected, string(data))
}

func TestLpaFindAttorneyIndex(t *testing.T) {
	lpa := Lpa{LpaInit: LpaInit{
		Attorneys: []Attorney{
			{Person: Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"}},
			{Person: Person{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3e"}},
		}},
	}

	testcases := map[string]struct {
		expectedIdx int
		ok          bool
	}{
		"9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d": {expectedIdx: 0, ok: true},
		"9ac5cb7c-fc75-40c7-8e53-059f36dbbe3e": {expectedIdx: 1, ok: true},
		"9ac5cb7c-fc75-40c7-8e53-059f36dbbe3f": {expectedIdx: 0, ok: false},
		"0":                                    {expectedIdx: 0, ok: true},
		"1":                                    {expectedIdx: 1, ok: true},
		"2":                                    {expectedIdx: 0, ok: false},
	}

	for changeKey, tc := range testcases {
		t.Run(changeKey, func(t *testing.T) {
			idx, ok := lpa.FindAttorneyIndex(changeKey)
			assert.Equal(t, tc.expectedIdx, idx)
			assert.Equal(t, tc.ok, ok)
		})
	}
}

func TestLpaFindTrustCorporationIndex(t *testing.T) {
	lpa := Lpa{LpaInit: LpaInit{
		TrustCorporations: []TrustCorporation{
			{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"},
			{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3e"},
		}},
	}

	testcases := map[string]struct {
		expectedIdx int
		ok          bool
	}{
		"9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d": {expectedIdx: 0, ok: true},
		"9ac5cb7c-fc75-40c7-8e53-059f36dbbe3e": {expectedIdx: 1, ok: true},
		"9ac5cb7c-fc75-40c7-8e53-059f36dbbe3f": {expectedIdx: 0, ok: false},
		"0":                                    {expectedIdx: 0, ok: true},
		"1":                                    {expectedIdx: 1, ok: true},
		"2":                                    {expectedIdx: 0, ok: false},
	}

	for changeKey, tc := range testcases {
		t.Run(changeKey, func(t *testing.T) {
			idx, ok := lpa.FindTrustCorporationIndex(changeKey)
			assert.Equal(t, tc.expectedIdx, idx)
			assert.Equal(t, tc.ok, ok)
		})
	}
}
