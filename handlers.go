package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spencer-p/surfdash/pkg/meta"
	"github.com/spencer-p/surfdash/pkg/noaa"
	"github.com/spencer-p/surfdash/pkg/sunset"
)

func fetchGoodTimes(numDays int) ([]meta.GoodTime, error) {
	query := noaa.PredictionQuery{
		Start:    time.Now(),
		Duration: time.Duration(numDays) * 24 * time.Hour,
		Station:  noaa.SantaCruz,
	}

	preds, err := noaa.GetPredictions(&query)
	if err != nil {
		return nil, err
	}

	sunevents := sunset.GetSunEvents(time.Now(), query.Duration, sunset.SantaCruz)

	goodTimes := meta.GoodTimes(meta.Conditions{preds, sunevents})
	return goodTimes, nil
}

func serveGoodTimes(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	goodTimes, err := fetchGoodTimes(7) // 7 days of forecast
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to get data: %+v", err)
		log.Printf("Failed to get data: %+v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	for _, gt := range goodTimes {
		fmt.Fprintf(w, "%s\n", gt.String())
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world\n")
}
