package splines

import (
	"fmt"
	"math"
	"time"

	"github.com/spencer-p/surfdash/pkg/noaa"
)

func ExampleDiscrete() {
	tstart := time.Date(2021, time.April, 3, 10, 30, 0, 0, time.Local)
	preds := noaa.Predictions{{
		Time:   noaa.Time(tstart),
		Height: 10,
	}, {
		Time:   noaa.Time(tstart.Add(1000 * time.Hour)),
		Height: 1,
	}}
	discrete := Discrete(CurvesBetween(preds), 10)
	for i := range discrete {
		fmt.Println(math.Round(discrete[i]))
	}
	// Output:
	// 10
	// 10
	// 9
	// 8
	// 6
	// 5
	// 3
	// 2
	// 1
	// 1
}
