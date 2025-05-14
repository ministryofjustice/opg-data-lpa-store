package event

type LpaUpdated struct {
	Uid        string `json:"uid"`
	ChangeType string `json:"changeType"`
}

type Metrics struct {
	Metrics []*Metric `json:"metrics"`
}

type Metric struct {
	Project          string
	Category         string
	Subcategory      string
	Environment      string
	MeasureName      string
	MeasureValue     string
	MeasureValueType string
	Time             string
}
