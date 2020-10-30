package meta

import (
	"fmt"
	"testing"
	"time"

	"github.com/spencer-p/surfdash/pkg/noaa"
	"github.com/spencer-p/surfdash/pkg/sunset"

	"github.com/google/go-cmp/cmp"
)

func date(d string) time.Time {
	t, err := time.Parse("01/02 3:04 PM", d)
	if err != nil {
		panic(err)
	}
	return t
}

func TestGoodTimes(t *testing.T) {
	table := map[string]struct {
		in   Conditions
		want []GoodTime
	}{
		"noop": {
			in: Conditions{
				Tides:     noaa.Predictions{},
				SunEvents: sunset.SunEvents{},
			},
			want: []GoodTime{},
		},
		"daytime low tide": {
			in: Conditions{
				Tides: noaa.Predictions{
					noaa.Prediction{
						Time:   noaa.Time(date("10/30 1:00 PM")),
						Height: 0.5,
						Type:   noaa.LowTide,
					},
				},
				SunEvents: sunset.SunEvents{
					sunset.SunEvent{
						Time:  date("10/30 7:00 AM"),
						Event: sunset.Sunrise,
					},
					sunset.SunEvent{
						Time:  date("10/30 6:00 PM"),
						Event: sunset.Sunset,
					},
				},
			},
			want: []GoodTime{
				GoodTime{
					Time:    date("10/30 1:00 PM"),
					Reasons: []string{tideReason(0.5)},
				},
			},
		},
		"low tide before sunrise": {
			in: Conditions{
				Tides: noaa.Predictions{
					noaa.Prediction{
						Time:   noaa.Time(date("10/30 6:00 AM")),
						Height: 0.5,
						Type:   noaa.LowTide,
					},
				},
				SunEvents: sunset.SunEvents{
					sunset.SunEvent{
						Time:  date("10/30 6:20 AM"),
						Event: sunset.Sunrise,
					},
					sunset.SunEvent{
						Time:  date("10/30 6:00 PM"),
						Event: sunset.Sunset,
					},
				},
			},
			want: []GoodTime{
				GoodTime{
					Time: date("10/30 6:00 AM"),
					Reasons: []string{
						tideReason(0.5),
						fmt.Sprintf("only %d minutes before sunrise", 20)},
				},
			},
		},
		"low tide after sunset": {
			in: Conditions{
				Tides: noaa.Predictions{
					noaa.Prediction{
						Time:   noaa.Time(date("10/30 6:20 PM")),
						Height: 0.5,
						Type:   noaa.LowTide,
					},
				},
				SunEvents: sunset.SunEvents{
					sunset.SunEvent{
						Time:  date("10/30 7:00 AM"),
						Event: sunset.Sunrise,
					},
					sunset.SunEvent{
						Time:  date("10/30 6:00 PM"),
						Event: sunset.Sunset,
					},
				},
			},
			want: []GoodTime{
				GoodTime{
					Time: date("10/30 6:20 PM"),
					Reasons: []string{
						tideReason(0.5),
						fmt.Sprintf("%d minutes after sunset", 20),
					},
				},
			},
		},
		"daytime low tide but it's high": {
			in: Conditions{
				Tides: noaa.Predictions{
					noaa.Prediction{
						Time:   noaa.Time(date("10/30 1:00 PM")),
						Height: 5.4,
						Type:   noaa.LowTide,
					},
				},
				SunEvents: sunset.SunEvents{
					sunset.SunEvent{
						Time:  date("10/30 7:00 AM"),
						Event: sunset.Sunrise,
					},
					sunset.SunEvent{
						Time:  date("10/30 6:00 PM"),
						Event: sunset.Sunset,
					},
				},
			},
			want: []GoodTime{},
		},
		"no sun events": {
			in: Conditions{
				Tides: noaa.Predictions{
					noaa.Prediction{
						Time:   noaa.Time(date("10/30 1:00 PM")),
						Height: 0.4,
						Type:   noaa.LowTide,
					},
				},
				SunEvents: sunset.SunEvents{},
			},
			want: []GoodTime{},
		},
		"sun events are not relevant": {
			in: Conditions{
				Tides: noaa.Predictions{
					noaa.Prediction{
						Time:   noaa.Time(date("10/30 1:00 PM")),
						Height: 0.4,
						Type:   noaa.LowTide,
					},
				},
				SunEvents: sunset.SunEvents{
					sunset.SunEvent{
						Time:  date("10/31 6:00 PM"),
						Event: sunset.Sunrise,
					},
				},
			},
			want: []GoodTime{},
		},
	}

	for name, tc := range table {
		t.Run(name, func(t *testing.T) {
			got := GoodTimes(tc.in)

			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("incorrect result: (-got,+want): %s", diff)
			}
		})
	}
}
