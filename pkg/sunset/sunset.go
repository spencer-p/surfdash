package sunset

import (
	"math"
	"time"

	"github.com/spencer-p/surfdash/pkg/timetricks"

	"github.com/keep94/sunrise"
)

const (
	twilightDur = 30 * time.Minute
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
	numDays := getDays(duration)
	ret := make(SunEvents, numDays*2)
	for i := 0; i < numDays*2; i += 2 {
		ret[i] = SunEvent{s.Sunrise(), Sunrise}
		ret[i+1] = SunEvent{s.Sunset(), Sunset}
		s.AddDays(1)
	}
	return ret
}

func getDays(t time.Duration) int {
	return int(math.Ceil(t.Hours() / 24))
}

// SunUp returns true if the sun is up at the given time.
// If the SunEvents provided cannot say, it returns false.
func (evs SunEvents) SunUp(t time.Time) bool {
	n := len(evs)
	left, right := 0, n
	// There must always be two elements (after rise and post set) to consider.
	for right-left >= 2 {
		// evs[mid] and evs[mid-1] are defined because there are at least two
		// elements.
		mid := left + (right-left)/2
		if t.After(evs[mid-1].Time) && t.Before(evs[mid].Time) {
			// Check if t falls between mid-1 and mid.
			// If those events are rise and set, sun is up.
			if evs[mid-1].Event == Sunrise &&
				evs[mid].Event == Sunset {
				return true
			} else {
				return false
			}
		} else if t.Before(evs[mid-1].Time) {
			right = mid
		} else {
			left = mid
		}
	}
	return false
}

// Dawn returns true if t is just before or at dawn.
func (evs SunEvents) Dawn(t time.Time) bool {
	return evs.SunUp(t.Add(twilightDur))
}

// Dusk returns true if t is just after or at dusk.
func (evs SunEvents) Dusk(t time.Time) bool {
	return evs.SunUp(t.Add(-twilightDur))
}
