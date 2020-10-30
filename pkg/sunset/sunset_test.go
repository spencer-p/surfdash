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
	// 26 Oct 20 07:27 PDT Sunrise
	// 26 Oct 20 18:18 PDT Sunset
	// 27 Oct 20 07:28 PDT Sunrise
	// 27 Oct 20 18:17 PDT Sunset
	// 28 Oct 20 07:29 PDT Sunrise
	// 28 Oct 20 18:16 PDT Sunset
	// 29 Oct 20 07:30 PDT Sunrise
	// 29 Oct 20 18:15 PDT Sunset
}
