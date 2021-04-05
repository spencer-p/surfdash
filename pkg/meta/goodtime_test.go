package meta

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/spencer-p/surfdash/pkg/timetricks"
)

func TestGoodTimeString(t *testing.T) {
	table := []struct {
		gt   GoodTime
		want string
	}{{
		gt: GoodTime{
			// seconds and nseconds should be unused
			Time:    time.Date(1999, time.January, 5, 5, 35, 20, 4, time.Local),
			Reasons: []string{"there is no kelp"},
		},
		want: "01/05 at 5:35 AM, there is no kelp",
	}, {
		gt: GoodTime{
			Time: timetricks.SetClock(time.Now(), 16, 27),
			Reasons: []string{
				"the sun is up",
				"you will be barreled",
			},
		},
		want: "Today at 4:27 PM, the sun is up and you will be barreled",
	}, {
		gt: GoodTime{
			Time: timetricks.SetClock(time.Now().Add(24*time.Hour), 12, 55),
			Reasons: []string{
				"the sun is up",
				"you will be barreled",
				"it's lunch time",
			},
		},
		want: "Tomorrow at 12:55 PM, the sun is up and you will be barreled and it's lunch time",
	}, {
		gt: GoodTime{
			// Set the time to three days from now so as not to trigger
			// today/tomorrow behavior.
			Time:    timetricks.SetClock(time.Now().Add(3*24*time.Hour), 13, 0),
			Reasons: []string{"the weather is nice"},
		},
		want: fmt.Sprintf("%s at 1:00 PM, the weather is nice", time.Now().Add(3*24*time.Hour).Weekday().String()),
	}}

	for _, tc := range table {
		t.Run(tc.want, func(t *testing.T) {
			got := tc.gt.String()
			if got != tc.want {
				t.Errorf("got %q, wanted %q", got, tc.want)
			}
		})
	}
}

func TestGoodTimeRoundTrip(t *testing.T) {
	gt := GoodTime{
		Time:    time.Date(1999, time.January, 5, 5, 35, 20, 4, time.Local),
		Reasons: []string{"there is no kelp"},
	}

	blob, err := json.Marshal(&gt)
	if err != nil {
		t.Errorf("unexpected: %v", err)
	}
	var got GoodTime
	if err := json.Unmarshal(blob, &got); err != nil {
		t.Errorf("unexpected: %v", err)
	}

	if diff := cmp.Diff(gt.String(), got.String()); diff != "" {
		t.Errorf("failed round trip (-want,+got):\n%s", diff)
	}
}
