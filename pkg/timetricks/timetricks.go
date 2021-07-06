package timetricks

import (
	"time"
)

const (
	dayFormat      = "20060102"
	weekPlusMinute = 7*24*time.Hour + time.Minute
)

// SameDay returns true if t and t2 represent the same calendar date.
func SameDay(t time.Time, t2 time.Time) bool {
	return t.Format(dayFormat) == t2.Format(dayFormat)
}

// Today returns true if t is today.
func Today(t time.Time) bool {
	return SameDay(t, time.Now())
}

// Tomorrow returns true if t is tomorrow.
func Tomorrow(t time.Time) bool {
	return Today(t.Add(-24 * time.Hour))
}

// TrimClock removes the wall clock time from t. The resulting time occurs on
// the same day at 00:00:00.00.
func TrimClock(t time.Time) time.Time {
	ns := t.Nanosecond()
	h, m, s := t.Clock()
	return t.Add(-1 *
		(time.Duration(h)*time.Hour +
			time.Duration(m)*time.Minute +
			time.Duration(s)*time.Second +
			time.Duration(ns)*time.Nanosecond))
}

// WithinWeek returns true if t occurs in the upcoming week from today.
func WithinWeek(t time.Time) bool {
	// Trim current time so they have no wall clock component, just
	// calendar date, and use it to compute the first minute of the coming week.
	// Then check if our time t occurs before then, as well as after the start
	// of today (minus a minute in case t falls at midnight).
	now := TrimClock(time.Now())
	firstMinuteOfNextWeek := now.Add(weekPlusMinute)
	return t.After(now.Add(-1*time.Minute)) && t.Before(firstMinuteOfNextWeek)

}

// SetClock sets the wall clock time of t to match the given hour, minute, and
// no seconds.
func SetClock(t time.Time, hour, minute time.Duration) time.Time {
	return TrimClock(t).Add(hour*time.Hour + minute*time.Minute)
}

// UniqueDay returns a string representation of t that is unique by the day.
// For instance, two seperate times on the same calendar day return identical
// strings.
func UniqueDay(t time.Time) string {
	return t.Format(dayFormat)
}
