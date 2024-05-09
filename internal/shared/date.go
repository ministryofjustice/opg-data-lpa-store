package shared

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Date struct {
	t           time.Time
	IsMalformed bool
}

func (d Date) IsZero() bool {
	return d.t.IsZero()
}

func (d *Date) UnmarshalJSON(data []byte) error {
	end := len(data) - 1
	if len(data) <= 2 || data[0] != '"' || data[end] != '"' {
		d.IsMalformed = len(data) != 0 && (len(data) != 2 || data[0] != '"' || data[end] != '"')
		return nil
	}

	if err := d.UnmarshalText(data[1:end]); err != nil {
		d.IsMalformed = true
	}

	return nil
}

func (d *Date) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	var err error
	d.t, err = time.Parse(time.DateOnly, string(data))
	return err
}

func (d Date) MarshalText() ([]byte, error) {
	if d.t.IsZero() {
		return nil, nil
	}

	return []byte(d.t.Format(time.DateOnly)), nil
}

func (d *Date) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	var s string
	if err := attributevalue.Unmarshal(av, &s); err != nil {
		return err
	}

	return d.UnmarshalText([]byte(s))
}

func (d Date) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	bytes, err := d.MarshalText()
	if err != nil {
		return nil, err
	}

	return attributevalue.Marshal(string(bytes))
}
