package meta

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/spencer-p/surfdash/pkg/noaa"
	"github.com/spencer-p/surfdash/pkg/sunset"
)

const (
	tideThresh       = 2.0 // feet
	firstLightThresh = 30 * time.Minute
)

var notFound = errors.New("not found")

// Conditions is the set of data we can perform meta analysis on.
type Conditions struct {
	Tides     noaa.Predictions
	SunEvents sunset.SunEvents
}

// GoodTimes analyzes a set of Conditions to find good times to surf.
func GoodTimes(c Conditions) []GoodTime {
	result := []GoodTime{}
	for _, tide := range c.Tides {
		// High tide is not interesting
		if tide.Type != noaa.LowTide {
			continue
		}

		// If the low tide is still pretty high, not interested
		if tide.Height > tideThresh {
			continue
		}

		// cast away the silly NOAA type
		t := time.Time(tide.Time)

		// Find last sun event that comes before the tide event
		suni, err := indexOfLastEventBefore(t, c.SunEvents)
		if err != nil {
			// No time before this event.
			// It is possible it happens before sunrise.
			if len(c.SunEvents) > 0 && c.SunEvents[0].Event == sunset.Sunrise {
				if gt, err := dawnPatrol(tide, c.SunEvents[0]); err == nil {
					result = append(result, gt)
				}
			}
			// Assuming there is not a "sunset first" case and the alternative
			// is no data.
			continue
		}

		if c.SunEvents[suni].Event == sunset.Sunset {
			// After sunset? Can't do that, unless ..
			if diff := t.Sub(c.SunEvents[suni].Time); diff < firstLightThresh {
				// Unless it's close to right after sunset
				result = append(result, GoodTime{
					Time: c.SunEvents[suni].Time,
					Reasons: []string{
						fmt.Sprintf("tide is low at %f", tide.Height),
						fmt.Sprintf("%.0f minutes after sunset", diff.Minutes()),
					},
				})

			} else if suni+1 < len(c.SunEvents) {
				// Check if sunrise is coming up..
				if gt, err := dawnPatrol(tide, c.SunEvents[suni+1]); err == nil {
					result = append(result, gt)
				}
			}
		} else if c.SunEvents[suni].Event == sunset.Sunrise {
			// Low tide during the day
			result = append(result, GoodTime{
				Time: t,
				Reasons: []string{
					fmt.Sprintf("tide is low at %f", tide.Height),
				},
			})
			continue
		}

	}

	return result
}

// dawnPatrol finds a GoodTime before dawn.
func dawnPatrol(tide noaa.Prediction, event sunset.SunEvent) (GoodTime, error) {
	t := time.Time(tide.Time)
	diff := event.Time.Sub(t)
	if diff > firstLightThresh {
		return GoodTime{}, notFound
	}
	return GoodTime{
		Time: t,
		Reasons: []string{
			fmt.Sprintf("tide is low at %f", tide.Height),
			fmt.Sprintf("only %.0f minutes before sunrise", diff.Minutes()),
		},
	}, nil
}

// Returns last event before time t, or an error if there is none.
func indexOfLastEventBefore(t time.Time, events sunset.SunEvents) (int, error) {
	// Remember, sort.Search finds the FIRST element. We have to reverse the
	// index.
	n := len(events)
	revi := sort.Search(n, func(revtesti int) bool {
		testi := n - 1 - revtesti
		return events[testi].Time.Before(t)
	})
	result := n - 1 - revi
	if result < 0 || result >= n {
		// no element found
		return -1, notFound
	}
	return result, nil
}
