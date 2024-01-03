package shared

import "encoding/json"

type Change struct {
	Key string          `json:"key"`
	Old json.RawMessage `json:"old"`
	New json.RawMessage `json:"new"`
}

type Update struct {
	Type    string   `json:"type"`
	Changes []Change `json:"changes"`
}
