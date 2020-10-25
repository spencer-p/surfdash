package meta

import (
	"time"
)

const (
	dayFormat = "20060102"
)

func isToday(t time.Time) bool {
	return t.Format(dayFormat) == time.Now().Format(dayFormat)
}

func isTomorrow(t time.Time) bool {
	return isToday(t.Add(-24 * time.Hour))
}

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
