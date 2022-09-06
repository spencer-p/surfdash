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

// Options specifies options to tune GoodTimes.
type Options struct {
	// When LowTideThresh and HighTideThresh are specified,
	// the resulting GoodTimes must have a tide level between
	// Low and High.
	LowTideThresh  *float64
	HighTideThresh *float64

	// Represents the default values. Optional.
	DefaultLowTide, DefaultHighTide *float64
}

// GoodTimes2 is like GoodTimes but better.
func GoodTimes2(c Conditions, opts Options) []GoodTime {
	opts.ApplyDefaults()
	result := []GoodTime{}
	preds := c.Tides

	tstart := time.Time(preds[0].Time)
	tend := time.Time(preds[len(preds)-1].Time)
	const step = 5 * time.Minute

	spl := splines.CurvesBetween(preds)

	for t := tstart; t.Before(tend); t = t.Add(step) {
		var gt GoodTime
		low := math.MaxFloat64
		var lowt time.Time
		for ; t.Before(tend); t = t.Add(step) {
			// If no desired tide, bail.
			tideHeight := spl.Eval(t)
			if tideHeight > *opts.HighTideThresh || tideHeight < *opts.LowTideThresh {
				break
			}

			// If the sun won't be shining, bail.
			if light := c.SunEvents.SunUp(t) || c.SunEvents.Dawn(t) || c.SunEvents.Dusk(t); !light {
				break
			}

			// Set the start time of this good time if needed and update the
			// duration to match.
			if gt.Time.IsZero() {
				gt.Time = t
			}
			gt.Duration = t.Sub(gt.Time) // + step

			// TODO Add reasons in a reasonable way.
			if tideHeight < low {
				low = tideHeight
				lowt = t
			}
		}
		if !gt.Time.IsZero() {
			if !gt.Time.Equal(lowt) {
				// The lowest part of good time is not the first time bucket.
				// This means we can specify the tide height at the start
				// without being redundant.
				gt.Reasons = append(gt.Reasons, fmt.Sprintf("tide is %.1fft at %s", noaa.Height(spl.Eval(gt.Time)), gt.Time.Format(timeFmt)))
			}
			gt.Reasons = append(gt.Reasons, fmt.Sprintf("tide is %.1fft at %s", noaa.Height(low), lowt.Format(timeFmt)))
			tend := gt.Time.Add(gt.Duration)
			if !tend.Equal(lowt) {
				// The lowest part is not the last time bucket.
				// Again, we can be more detailed without being redundant.
				gt.Reasons = append(gt.Reasons, fmt.Sprintf("tide is %.1fft at %s", noaa.Height(spl.Eval(tend)), tend.Format(timeFmt)))
			}
			result = append(result, gt)
		}
	}
	return result
}

func (o *Options) ApplyDefaults() {
	if o.LowTideThresh == nil {
		low := float64(-1000)
		o.LowTideThresh = &low
	}
	if o.HighTideThresh == nil {
		high := float64(smallTideThresh)
		o.HighTideThresh = &high
	}
}
