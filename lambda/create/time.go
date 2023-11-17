package main

import "time"

type Time struct {
	time.Time
	Err error
}

func (m *Time) UnmarshalJSON(data []byte) error {
	err := m.Time.UnmarshalJSON(data)

	*m = Time{
		Time: m.Time,
		Err:  err,
	}

	return nil
}
