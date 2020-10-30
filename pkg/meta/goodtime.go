package meta

import (
	"fmt"
	"strings"
	"time"
)

const (
	dayFmt  = "01/02"
	timeFmt = "3:04 PM"
)

// GoodTime represents a good time to go surfing.
type GoodTime struct {
	Time    time.Time
	Reasons []string
}

func (gt *GoodTime) String() string {
	var day string
	if isToday(gt.Time) {
		day = "Today"
	} else if isTomorrow(gt.Time) {
		day = "Tomorrow"

	} else if withinWeek(gt.Time) {
		day = gt.Time.Weekday().String()
	} else {
		day = gt.Time.Format(dayFmt)
	}

	return fmt.Sprintf("%s at %s, %s",
		day,
		gt.Time.Format(timeFmt),
		strings.Join(gt.Reasons, " and "))
}
