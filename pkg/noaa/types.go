package noaa

import "time"

type Prediction struct {
	Time   string `json:"t"` // TODO make this field time.Time
	Height string `json:"v"`
	Type   string `json:"type"` // "H" or "L"
}

type PredictionList []Prediction

type Predictions struct {
	Predictions PredictionList `json:"predictions"`
}

type PredictionQuery struct {
	Start, End time.Time
	Station    int
}
