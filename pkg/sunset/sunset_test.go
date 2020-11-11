package sunset

import (
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

func ExampleGetSunEvents() {
	start := time.Date(2020, time.October, 30, 10, 18, 0, 0, SantaCruz.Location)
	dur := 5 * 24 * time.Hour
	events := GetSunEvents(start, dur, SantaCruz)
	for _, e := range events {
		fmt.Printf("%s\n", e.String())
	}
	// Output:
	// 30 Oct 20 07:31 PDT Sunrise
	// 30 Oct 20 18:13 PDT Sunset
	// 31 Oct 20 07:33 PDT Sunrise
	// 31 Oct 20 18:12 PDT Sunset
	// 01 Nov 20 06:34 PST Sunrise
	// 01 Nov 20 17:11 PST Sunset
	// 02 Nov 20 06:35 PST Sunrise
	// 02 Nov 20 17:10 PST Sunset
	// 03 Nov 20 06:36 PST Sunrise
	// 03 Nov 20 17:09 PST Sunset
}

func TestGetDays(t *testing.T) {
	f := func(want int) bool {
		if want > 1e10 || want < 0 {
			// skip unreasonably high values
			return true
		}

		input := time.Duration(want) * 24 * time.Hour
		got := getDays(input)
		return want == got
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
