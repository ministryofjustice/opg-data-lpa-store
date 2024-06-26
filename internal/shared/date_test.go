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
	testcases := map[string]struct {
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
		"double char not a string": {
			in:                `11`,
			expectFormatted:   "1 January 0001",
			expectIsMalformed: true,
		},
		"space": {
			in:                `" "`,
			expectFormatted:   "1 January 0001",
			expectIsMalformed: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			var date Date
			err := json.Unmarshal([]byte(tc.in), &date)

			assert.Nil(t, err)
			assert.Equal(t, tc.expectFormatted, date.t.Format("2 January 2006"))
			assert.Equal(t, tc.expectIsMalformed, date.IsMalformed)

			if !tc.expectIsMalformed {
				marshal, err := json.Marshal(date)
				assert.Nil(t, err)
				assert.Equal(t, tc.in, string(marshal))
			}
		})
	}
}

func TestDateDynamoDB(t *testing.T) {
	testcases := map[string]struct {
		dynamo string
		date   Date
	}{
		"value": {
			dynamo: "2000-01-02",
			date:   Date{t: time.Date(2000, time.January, 2, 0, 0, 0, 0, time.UTC)},
		},
		"zero": {
			dynamo: "",
			date:   Date{},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			av := &types.AttributeValueMemberS{Value: tc.dynamo}

			var unmarshal Date
			assert.Nil(t, attributevalue.Unmarshal(av, &unmarshal))
			assert.Equal(t, tc.date, unmarshal)

			marshal, err := attributevalue.Marshal(unmarshal)
			assert.Nil(t, err)
			assert.Equal(t, av, marshal)
		})
	}
}
