package shared

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalDate(t *testing.T) {
	testCases := map[string]struct {
		name              string
		in                string
		expectFormatted   string
		expectIsMalformed bool
	}{
		"ok": {
			in:                `"1930-10-31"`,
			expectFormatted:   "31 October 1930",
			expectIsMalformed: false,
		},
		"out of bounds": {
			in:                `"1930-11-31"`,
			expectFormatted:   "1 January 0001",
			expectIsMalformed: true,
		},
		"invalid string": {
			in:                `"31 October 1930"`,
			expectFormatted:   "1 January 0001",
			expectIsMalformed: true,
		},
		"number": {
			in:                `1700240133`,
			expectFormatted:   "1 January 0001",
			expectIsMalformed: true,
		},
		"empty string": {
			in:                `""`,
			expectFormatted:   "1 January 0001",
			expectIsMalformed: false,
		},
		"space": {
			in:                `" "`,
			expectFormatted:   "1 January 0001",
			expectIsMalformed: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var date Date
			err := json.Unmarshal([]byte(tc.in), &date)

			assert.Nil(t, err)
			assert.Equal(t, tc.expectFormatted, date.Time.Format("2 January 2006"))
			assert.Equal(t, tc.expectIsMalformed, date.IsMalformed)
		})
	}
}

func TestDateDynamoDB(t *testing.T) {
	in := &types.AttributeValueMemberS{Value: "2000-01-02"}

	var v Date
	attributevalue.Unmarshal(in, &v)
	assert.Equal(t, Date{Time: time.Date(2000, time.January, 2, 0, 0, 0, 0, time.UTC)}, v)

	out, _ := attributevalue.Marshal(v)
	assert.Equal(t, in, out)
}
