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

type Update struct {
	Id      string   `json:"id"`      // UUID for the update
	Uid     string   `json:"uid"`     // UID of the changed LPA
	Applied string   `json:"applied"` // RFC3339 datetime
	Author  string   `json:"author"`
	Type    string   `json:"type"`
	Changes []Change `json:"changes"`
}

func (u Update) AuthorUID() string {
	parts := strings.Split(u.Author, ":users:")
	if len(parts) > 1 && parts[1] != "" {
		return parts[1]
	} else {
		return ""
	}
}
