package meta

import (
	"testing"
	"time"
)

func trimClock(t time.Time) time.Time {
	h, m, s := t.Clock()
	return t.Add(-1 *
		(time.Duration(h)*time.Hour +
			time.Duration(m)*time.Minute +
			time.Duration(s)*time.Second))
}

func setClock(t time.Time, hour, minute time.Duration) time.Time {
	return trimClock(t).Add(hour*time.Hour + minute*time.Minute)
}

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
			Time:    setClock(time.Now(), 16, 27),
			Reasons: []string{"the sun is up", "you will be barreled"},
		},
		want: "today at 4:27 PM, the sun is up and you will be barreled",
	}, {
		gt: GoodTime{
			Time:    setClock(time.Now().Add(24*time.Hour), 12, 55),
			Reasons: []string{"the sun is up", "you will be barreled", "it's lunch time"},
		},
		want: "tomorrow at 12:55 PM, the sun is up and you will be barreled and it's lunch time",
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
