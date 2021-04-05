package main

import (
	"fmt"
	"time"

	"github.com/spencer-p/surfdash/pkg/noaa"
	"github.com/spencer-p/surfdash/pkg/noaa/splines"
)

func main() {
	dur := 14 * 24 * time.Hour
	step := 2 * time.Hour

	query := noaa.PredictionQuery{
		Start:    time.Now(),
		Duration: dur,
		Station:  noaa.SantaCruz,
	}

	preds, err := noaa.GetPredictions(&query)
	if err != nil {
		fmt.Printf("failed to fetch from NOAA: %v\n", err)
		return
	}

	tstart := time.Time(preds[0].Time)
	tend := tstart.Add(dur)
	spl := splines.CurvesBetween(preds)
	for t := tstart; t.Before(tend); t = t.Add(step) {
		fmt.Printf("%f ", spl.Eval(t))
	}
}
