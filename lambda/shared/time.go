package shared

import (
	"regexp"
	"time"
)

type Time struct {
	time.Time
	IsMalformed bool
}

func (m *Time) UnmarshalJSON(data []byte) error {
	time, err := time.Parse(`"2006-01-02"`, string(data))

	m.Time = time

	if err != nil {
		if ok, _ := regexp.MatchString("cannot parse \".+\" as \"\\\"2006-01-02\\\"\"$", err.Error()); ok {
			m.IsMalformed = true
		}
	}

	return nil
}
