package shared

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLpaInitMarshalJSON(t *testing.T) {
	expected := `{
"lpaType":"",
"donor":{"firstNames":"","lastName":"","address":{"line1":"","line2":"","line3":"","town":"","postcode":"","country":""},"dateOfBirth":"0001-01-01T00:00:00Z","email":"","otherNamesKnownBy":""},
"attorneys":null,
"certificateProvider":{"firstNames":"","lastName":"","address":{"line1":"","line2":"","line3":"","town":"","postcode":"","country":""},"email":"","phone":"","channel":"","signedAt":"0001-01-01T00:00:00Z"},
"signedAt":"0001-01-01T00:00:00Z"
}`

	data, _ := json.Marshal(LpaInit{})
	assert.JSONEq(t, expected, string(data))
}
