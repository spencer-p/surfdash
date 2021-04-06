package meta

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/spencer-p/surfdash/pkg/noaa"
	"github.com/spencer-p/surfdash/pkg/noaa/splines"
	"github.com/spencer-p/surfdash/pkg/sunset"
)

const (
	tideThresh       = 2.0 // feet
	smallTideThresh  = 1.0 // feet
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
					Time: t,
					Reasons: []string{
						tideReason(tide.Height),
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
					tideReason(tide.Height),
				},
			})
			continue
		}

	}

	return result
}

func tideReason(height noaa.Height) string {
	return fmt.Sprintf("tide is low at %.2fft", height)
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
			tideReason(tide.Height),
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

// GoodTimes2 is like GoodTimes but better.
// TODO Name, document
func GoodTimes2(c Conditions) []GoodTime {
	result := []GoodTime{}
	preds := c.Tides

	tstart := time.Time(preds[0].Time)
	tend := time.Time(preds[len(preds)-1].Time)
	const step = 30 * time.Minute

	spl := splines.CurvesBetween(preds)

	for t := tstart; t.Before(tend); t = t.Add(step) {
		var gt GoodTime
		low := math.MaxFloat64
		for ; t.Before(tend); t = t.Add(step) {
			tideHeight := spl.Eval(t)
			var (
				sunup = c.SunEvents.SunUp(t)
				dawn  = c.SunEvents.Dawn(t)
				dusk  = c.SunEvents.Dusk(t)
			)
			// If no low tide or sun, bail.
			if tideHeight > smallTideThresh {
				break
			} else if !(sunup || dawn || dusk) {
				break
			}

			// Set the start time of this good time if needed and update the
			// duration to match.
			if gt.Time.IsZero() {
				gt.Time = t
			}
			gt.Duration = t.Sub(gt.Time) + step

			// TODO Add reasons in a reasonable way.
			if tideHeight < low {
				low = tideHeight
			}
		}
		if !gt.Time.IsZero() {
			gt.Reasons = append(gt.Reasons, tideReason(noaa.Height(low)))
			result = append(result, gt)
		}
	}
	return result
}
