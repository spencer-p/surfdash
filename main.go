package main

import (
	"fmt"
	"time"

	"github.com/spencer-p/surfdash/pkg/noaa"
)

func main() {
	fmt.Println("vim-go")

	query := noaa.PredictionQuery{
		Start:   time.Now(),
		End:     time.Now().Add(24 * time.Hour),
		Station: 9413745,
	}

	preds, err := noaa.GetPredictions(&query)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Printf("%+v\n", preds)
}
