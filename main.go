package main

import (
	"fmt"
	"time"

	"github.com/spencer-p/surfdash/pkg/meta"
	"github.com/spencer-p/surfdash/pkg/noaa"
	"github.com/spencer-p/surfdash/pkg/sunset"
)

func main() {
	query := noaa.PredictionQuery{
		Start:    time.Now(),
		Duration: 14 * 24 * time.Hour,
		Station:  noaa.SantaCruz,
	}

	preds, err := noaa.GetPredictions(&query)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	sunevents := sunset.GetSunEvents(time.Now(), query.Duration, sunset.SantaCruz)

	goodTimes := meta.GoodTimes(meta.Conditions{preds, sunevents})

	for _, gt := range goodTimes {
		fmt.Printf("%s\n", gt.String())
	}
}
