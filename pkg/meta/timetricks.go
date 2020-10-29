package meta

import (
	"time"
)

const (
	dayFormat      = "20060102"
	weekPlusMinute = 7*24*time.Hour + time.Minute
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

func withinWeek(t time.Time) bool {
	// Trim current time so they have no wall clock component, just
	// calendar date, and use it to compute the first minute of the coming week.
	// Then check if our time t occurs before then, as well as after the start
	// of today (minus a minute in case t falls at midnight).
	now := trimClock(time.Now())
	firstMinuteOfNextWeek := now.Add(weekPlusMinute)
	return t.After(now.Add(-1*time.Minute)) && t.Before(firstMinuteOfNextWeek)

}

func setClock(t time.Time, hour, minute time.Duration) time.Time {
	return trimClock(t).Add(hour*time.Hour + minute*time.Minute)
}
