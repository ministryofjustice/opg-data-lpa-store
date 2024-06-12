package shared

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLpaInitMarshalJSON(t *testing.T) {
	expected := `{
"lpaType":"","channel":"",
"donor":{"uid":"","firstNames":"","lastName":"","address":{"line1":"","country":""},"dateOfBirth":"","email":"","contactLanguagePreference":"",
"identityCheck":{"date":"0001-01-01T00:00:00Z","reference":"","type":""}},
"attorneys":null,
"certificateProvider":{"uid":"","firstNames":"","lastName":"","address":{"line1":"","country":""},"email":"","phone":"","channel":"",
"identityCheck":{"date":"0001-01-01T00:00:00Z","reference":"","type":""}},
"signedAt":"0001-01-01T00:00:00Z"
}`

	data, _ := json.Marshal(LpaInit{})
	assert.JSONEq(t, expected, string(data))
}
