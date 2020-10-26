package sunset

import (
	"fmt"
	"time"
)

func ExampleGetSunEvents() {
	start := time.Date(2020, time.October, 25, 0, 0, 0, 0, SantaCruz.Location)
	dur := 5 * 24 * time.Hour
	events := GetSunEvents(start, dur, SantaCruz)
	for _, e := range events {
		fmt.Printf("%s\n", e.String())
	}
	// Output:
	// 25 Oct 20 07:26 PDT Sunrise
	// 25 Oct 20 18:19 PDT Sunset
	// 25 Oct 20 07:26 PDT Sunrise
	// 25 Oct 20 18:19 PDT Sunset
	// 25 Oct 20 07:26 PDT Sunrise
	// 25 Oct 20 18:19 PDT Sunset
	// 25 Oct 20 07:26 PDT Sunrise
	// 25 Oct 20 18:19 PDT Sunset
	// 25 Oct 20 07:26 PDT Sunrise
	// 25 Oct 20 18:19 PDT Sunset
}
