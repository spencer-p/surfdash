package meta

import (
	"fmt"
	"strings"
	"time"

	"github.com/spencer-p/surfdash/pkg/noaa"
)

const (
	dayFmt  = "01/02"
	timeFmt = "3:04 PM"
)

// Conditions is the set of data we can perform meta analysis on.
type Conditions struct {
	Tides noaa.Predictions
}

// GoodTime represents a good time to go surfing.
type GoodTime struct {
	Time    time.Time
	Reasons []string
}

func (gt *GoodTime) String() string {
	var day string
	if isToday(gt.Time) {
		day = "today"
	} else if isTomorrow(gt.Time) {
		day = "tomorrow"
	} else {
		day = gt.Time.Format(dayFmt)
	}

	return fmt.Sprintf("%s at %s, %s",
		day,
		gt.Time.Format(timeFmt),
		strings.Join(gt.Reasons, " and "))
}

// GoodTimes analyzes a set of Conditions to find good times to surf.
func GoodTimes(c *Conditions) []GoodTime {
	return []GoodTime{}
}
