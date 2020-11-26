package noaa

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

const predTimeFormat = "2006-01-02 15:04"

// Prediction holds a single tide event prediction.
type Prediction struct {
	// Local time of tide prediction
	Time Time `json:"t"`
	// Height in feet
	Height Height `json:"v"`
	// High or Low tide, "H" or "L" when encoded
	Type Tide `json:"type"`
}

// Verify the custom types can be unmarshaled
var _ json.Unmarshaler = &Time{}
var _ json.Unmarshaler = new(Height)
var _ json.Unmarshaler = new(Tide)

// Predictions is a time series of Prediction.
type Predictions []Prediction

// NOAAResult is the data type returned by the NOAA API.
type NOAAResult struct {
	Predictions Predictions `json:"predictions"`
}

// PredictionQuery is used to query tide data at a station in a given time
// window; see GetPredictions.
type PredictionQuery struct {
	Start    time.Time
	Duration time.Duration
	Station  Station
}

type Station int

const (
	SantaCruz Station = 9413745
)

type Time time.Time

func (t *Time) UnmarshalJSON(buf []byte) error {
	var s string
	if err := json.Unmarshal(buf, &s); err != nil {
		return fmt.Errorf("prediction time %q not string: %w", buf, err)
	}
	parsed, err := time.ParseInLocation(predTimeFormat, s, time.Local)
	if err != nil {
		return fmt.Errorf("prediction time %q not in fmt %q: %w", s, predTimeFormat, err)
	}
	*t = Time(parsed)
	return nil
}

type Height float64

func (h *Height) UnmarshalJSON(buf []byte) error {
	var s string
	if err := json.Unmarshal(buf, &s); err != nil {
		return fmt.Errorf("water height %q not string: %w", buf, err)
	}
	parsed, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("water height %q not a float: %w", s, err)
	}
	*h = Height(parsed)
	return nil
}

type Tide uint

const (
	HighTide Tide = iota
	LowTide
)

func (t Tide) Valid() bool {
	return t == HighTide || t == LowTide
}

func (t *Tide) UnmarshalJSON(buf []byte) error {
	var s string
	if err := json.Unmarshal(buf, &s); err != nil {
		return fmt.Errorf("tide %q not a string: %w", buf, err)
	}
	switch s {
	case "H":
		*t = HighTide
	case "L":
		*t = LowTide
	default:
		return fmt.Errorf("invalid tide type %q", s)
	}
	return nil
}

func (t Tide) String() string {
	switch t {
	case HighTide:
		return "H"
	case LowTide:
		return "L"
	default:
		return "invalid"
	}
}

func (p Prediction) String() string {
	return fmt.Sprintf("{t: %s, v: %f, type: %s}",
		time.Time(p.Time).Format(time.RFC822),
		p.Height,
		p.Type.String())
}
