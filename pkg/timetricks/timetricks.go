package timetricks

import (
	"time"
)

const (
	dayFormat      = "20060102"
	weekPlusMinute = 7*24*time.Hour + time.Minute
)

func SameDay(t time.Time, t2 time.Time) bool {
	return t.Format(dayFormat) == t2.Format(dayFormat)
}

func Today(t time.Time) bool {
	return SameDay(t, time.Now())
}

func Tomorrow(t time.Time) bool {
	return Today(t.Add(-24 * time.Hour))
}

func TrimClock(t time.Time) time.Time {
	h, m, s := t.Clock()
	return t.Add(-1 *
		(time.Duration(h)*time.Hour +
			time.Duration(m)*time.Minute +
			time.Duration(s)*time.Second))
}

func WithinWeek(t time.Time) bool {
	// Trim current time so they have no wall clock component, just
	// calendar date, and use it to compute the first minute of the coming week.
	// Then check if our time t occurs before then, as well as after the start
	// of today (minus a minute in case t falls at midnight).
	now := TrimClock(time.Now())
	firstMinuteOfNextWeek := now.Add(weekPlusMinute)
	return t.After(now.Add(-1*time.Minute)) && t.Before(firstMinuteOfNextWeek)

}

func SetClock(t time.Time, hour, minute time.Duration) time.Time {
	return TrimClock(t).Add(hour*time.Hour + minute*time.Minute)
}

// UniqueDay returns a string representation of t that is unique by the day.
// For instance, two seperate times on the same calendar day return identical
// strings.
func UniqueDay(t time.Time) string {
	return t.Format(dayFormat)
}
