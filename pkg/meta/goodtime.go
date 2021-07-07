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
	day := timetricks.Day(gt.Time)

	until := ""
	if gt.Duration != 0 {
		until = fmt.Sprintf(" until %s", gt.Time.Add(gt.Duration).Format(timeFmt))
	}

	return fmt.Sprintf("%s at %s%s",
		day,
		gt.Time.Format(timeFmt),
		until)
}

// UpdatePrettyTime makes sure that the good time's pretty time is set.
func (gt *GoodTime) UpdatePrettyTime() {
	if gt.PrettyTime == "" {
		gt.PrettyTime = gt.prettyTime()
	}
}

// TimeRange returns a time range for the goodtime, similar to PrettyTime
// without the date.
func (gt *GoodTime) TimeRange() string {
	until := ""
	if gt.Duration != 0 {
		until = fmt.Sprintf(" until %s", gt.Time.Add(gt.Duration).Format(timeFmt))
	}
	return fmt.Sprintf("%s%s", gt.Time.Format(timeFmt), until)
}

func (gt *GoodTime) MarshalJSON() ([]byte, error) {
	// Fill in pretty time if needed.
	gt.UpdatePrettyTime()
	// Dereference is necessary to avoid infinite loop; this method
	// only has pointer receiver.
	return json.Marshal(*gt)
}
