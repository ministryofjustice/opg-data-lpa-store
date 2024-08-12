package shared

import (
	"encoding/json"
	"strings"
)

type Change struct {
	Key string          `json:"key"`
	Old json.RawMessage `json:"old"`
	New json.RawMessage `json:"new"`
}

type URN string

func (u URN) Details() AuthorDetails {
	parts := strings.Split(string(u), ":")

	if len(parts) != 6 || parts[3] == "" || parts[5] == "" {
		return AuthorDetails{}
	}

	return AuthorDetails{
		UID:     parts[5],
		Service: parts[3],
	}
}

type Update struct {
	Id      string   `json:"id"`      // UUID for the update
	Uid     string   `json:"uid"`     // UID of the changed LPA
	Applied string   `json:"applied"` // RFC3339 datetime
	Author  URN      `json:"author"`
	Type    string   `json:"type"`
	Changes []Change `json:"changes"`
}

type AuthorDetails struct {
	UID     string
	Service string
}
