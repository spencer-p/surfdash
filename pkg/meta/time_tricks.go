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
