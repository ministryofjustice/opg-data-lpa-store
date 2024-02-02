package shared

import "encoding/json"

type Change struct {
	Key string          `json:"key"`
	Old json.RawMessage `json:"old"`
	New json.RawMessage `json:"new"`
}

type Update struct {
	Uid     string   `json:"author"`
	Applied string   `json:"applied"` // RFC3339 datetime
	Author  string   `json:"author"`
	Type    string   `json:"type"`
	Changes []Change `json:"changes"`
}
