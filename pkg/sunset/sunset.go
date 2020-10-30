package sunset

import (
	"math"
	"time"

	"github.com/spencer-p/surfdash/pkg/timetricks"

	"github.com/keep94/sunrise"
)

// GetSunEvents returns a list of ordered sun events from the starting time to
// the end time in the given place. The first result will always be a sunrise.
func GetSunEvents(start time.Time, duration time.Duration, place Place) SunEvents {
	var s sunrise.Sunrise
	s.Around(place.Lat, place.Long, start)

	// Make sure we start with the correct day
	// The sunrise package is not very clean with its dates.
	// TODO Surely this breaks sometimes?
	for !timetricks.SameDay(start, s.Sunrise()) {
		s.AddDays(1)
	}

	// Get sunrises and sunsets for the given number of days.
	numDays := int(math.Ceil(duration.Hours() / 24))
	ret := make(SunEvents, numDays*2)
	for i := 0; i < numDays*2; i += 2 {
		ret[i] = SunEvent{s.Sunrise(), Sunrise}
		ret[i+1] = SunEvent{s.Sunset(), Sunset}
		s.AddDays(1)
	}
	return ret
}
