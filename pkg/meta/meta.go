package meta

import (
	"fmt"
	"time"

	"github.com/spencer-p/surfdash/pkg/noaa"
	"github.com/spencer-p/surfdash/pkg/sunset"
)

const tideThresh = 2.0 // feet

// Conditions is the set of data we can perform meta analysis on.
type Conditions struct {
	Tides     noaa.Predictions
	SunEvents sunset.SunEvents
}

// GoodTimes analyzes a set of Conditions to find good times to surf.
func GoodTimes(c Conditions) []GoodTime {
	result := []GoodTime{}
	suni := 0
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

		// Fast forward the sun events so that tide is comes after an event
		for {
			if suni >= len(c.SunEvents) {
				// out of sun events.
				return result
			}
			if t.After(c.SunEvents[suni].Time) {
				// condition met to exit loop - tide is now after a sun event
				break
			}
			suni += 1
		}

		// After sunset? Can't do that
		if c.SunEvents[suni].Event == sunset.Sunset {
			// Unless it's within half an hour TODO
			continue
		}

		// Low tide during the day
		if c.SunEvents[suni].Event == sunset.Sunrise {
			result = append(result, GoodTime{
				Time: t,
				Reasons: []string{
					fmt.Sprintf("tide is low at %f", tide.Height),
				},
			})
		}
	}

	return result
}
