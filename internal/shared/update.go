package shared

type Change struct {
	Key string      `json:"key"`
	Old interface{} `json:"old"`
	New interface{} `json:"new"`
}

type Update struct {
	Type    string   `json:"type"`
	Changes []Change `json:"changes"`
}
