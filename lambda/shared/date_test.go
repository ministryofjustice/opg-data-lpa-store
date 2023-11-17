package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalDate(t *testing.T) {
	testCases := []struct {
		name              string
		in                string
		expectFormatted   string
		expectIsMalformed bool
	}{
		{
			name:              "ok",
			in:                `"1930-10-31"`,
			expectFormatted:   "31 October 1930",
			expectIsMalformed: false,
		},
		{
			name:              "out of bounds",
			in:                `"1930-11-31"`,
			expectFormatted:   "1 January 0001",
			expectIsMalformed: true,
		},
		{
			name:              "invalid string",
			in:                `"31 October 1930"`,
			expectFormatted:   "1 January 0001",
			expectIsMalformed: true,
		},
		{
			name:              "number",
			in:                `1700240133`,
			expectFormatted:   "1 January 0001",
			expectIsMalformed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			date := Date{}
			date.UnmarshalJSON([]byte(tc.in))

			assert.Equal(t, tc.expectFormatted, date.Time.Format("2 January 2006"))
			assert.Equal(t, tc.expectIsMalformed, date.IsMalformed)
		})
	}
}
