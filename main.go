package main

import (
	"fmt"
	"time"

	"github.com/spencer-p/surfdash/pkg/noaa"
)

func main() {
	query := noaa.PredictionQuery{
		Start:   time.Now(),
		End:     time.Now().Add(2 * 24 * time.Hour),
		Station: noaa.SantaCruz,
	}

	preds, err := noaa.GetPredictions(&query)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	for _, pred := range preds {
		fmt.Printf("%+v\n", pred)
	}
}
