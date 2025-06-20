package event

type LpaUpdated struct {
	Uid        string `json:"uid"`
	ChangeType string `json:"changeType"`
}

type metrics struct {
	Metrics []metricWrapper `json:"metrics"`
}

type metricWrapper struct {
	Metric *Metric `json:"metric"`
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
