package shared

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLpaInitMarshalJSON(t *testing.T) {
	expected := `{
"lpaType":"","channel":"","language":"",
"donor":{"uid":"","firstNames":"","lastName":"","address":{"line1":"","country":""},"dateOfBirth":"","email":"","contactLanguagePreference":""},
"attorneys":null,
"certificateProvider":{"uid":"","firstNames":"","lastName":"","address":{"line1":"","country":""},"email":"","phone":"","channel":""},
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

	idx, ok := lpa.FindAttorneyIndex("9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d")
	assert.Equal(t, 0, idx)
	assert.True(t, ok)

	idx, ok = lpa.FindAttorneyIndex("9ac5cb7c-fc75-40c7-8e53-059f36dbbe3e")
	assert.Equal(t, 1, idx)
	assert.True(t, ok)

	idx, ok = lpa.FindAttorneyIndex("9ac5cb7c-fc75-40c7-8e53-059f36dbbe3f")
	assert.Equal(t, 0, idx)
	assert.False(t, ok)
}

func TestLpaFindTrustCorporationIndex(t *testing.T) {
	lpa := Lpa{LpaInit: LpaInit{
		TrustCorporations: []TrustCorporation{
			{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d"},
			{UID: "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3e"},
		}},
	}

	idx, ok := lpa.FindTrustCorporationIndex("9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d")
	assert.Equal(t, 0, idx)
	assert.True(t, ok)

	idx, ok = lpa.FindTrustCorporationIndex("9ac5cb7c-fc75-40c7-8e53-059f36dbbe3e")
	assert.Equal(t, 1, idx)
	assert.True(t, ok)

	idx, ok = lpa.FindTrustCorporationIndex("9ac5cb7c-fc75-40c7-8e53-059f36dbbe3f")
	assert.Equal(t, 0, idx)
	assert.False(t, ok)
}
