package event

type LpaUpdated struct {
	Uid        string `json:"uid"`
	ChangeType string `json:"changeType"`
}
