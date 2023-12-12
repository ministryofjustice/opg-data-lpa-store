package shared

import (
	"time"
)

type Date struct {
	time.Time
	IsMalformed bool
}

func (m *Date) UnmarshalJSON(data []byte) error {
	str := string(data)
	time, err := time.Parse(`"2006-01-02"`, str)

	m.Time = time

	if err != nil {
		m.IsMalformed = str != "" && str != `""`
	}

	return nil
}
