package meta

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spencer-p/surfdash/pkg/timetricks"
)

const (
	dayFmt  = "01/02"
	timeFmt = "3:04 PM"
)

// GoodTime represents a good time to go surfing.
type GoodTime struct {
	Time     time.Time     `json:unix_time`
	Reasons  []string      `json:"reasons"`
	Duration time.Duration `json:"duration",omitempty`

	// PrettyTime is a human-readable version of the time, relative to the
	// current date. Optional.
	PrettyTime string `json:"pretty_time",omitempty`
}

func (gt *GoodTime) String() string {
	return fmt.Sprintf("%s, %s",
		gt.prettyTime(),
		strings.Join(gt.Reasons, " and "))
}

func (gt *GoodTime) prettyTime() string {
	var day string
	if timetricks.Today(gt.Time) {
		day = "Today"
	} else if timetricks.Tomorrow(gt.Time) {
		day = "Tomorrow"
	} else if timetricks.WithinWeek(gt.Time) {
		day = gt.Time.Weekday().String()
	} else {
		day = gt.Time.Format(dayFmt)
	}

	return fmt.Sprintf("%s at %s",
		day,
		gt.Time.Format(timeFmt))
}

func (gt *GoodTime) MarshalJSON() ([]byte, error) {
	// Fill in pretty time if needed.
	if gt.PrettyTime == "" {
		gt.PrettyTime = gt.prettyTime()
	}
	// Dereference is necessary to avoid infinite loop; this method
	// only has pointer receiver.
	return json.Marshal(*gt)
}
