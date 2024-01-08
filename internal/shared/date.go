package shared

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Date struct {
	time.Time
	IsMalformed bool
}

func (d *Date) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || (len(data) == 2 && data[0] == '"' && data[1] == '"') {
		d.IsMalformed = true
		return nil
	}

	return d.UnmarshalText(data[1 : len(data)-1])
}

func (d *Date) UnmarshalText(data []byte) error {
	var err error
	d.Time, err = time.Parse(time.DateOnly, string(data))
	return err
}

func (d *Date) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	var s string
	if err := attributevalue.Unmarshal(av, &s); err != nil {
		return err
	}

	return d.UnmarshalText([]byte(s))
}

func (d Date) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	text := ""
	if !d.IsZero() {
		text = d.Time.Format(time.DateOnly)
	}

	return attributevalue.Marshal(text)
}
