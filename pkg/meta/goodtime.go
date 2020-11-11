package meta

import (
	"bytes"
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
	Time    time.Time
	Reasons []string
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
	reasons, err := json.Marshal(gt.Reasons)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "{\"pretty_time\": \"%s\", \"unix_time\": %d, \"reasons\": %s}",
		gt.prettyTime(),
		gt.Time.Unix(),
		reasons)
	return buf.Bytes(), nil
}

// used for unmarshaling tricks
type altGoodTime struct {
	UnixTime int64     `json:"unix_time"`
	Reasons  *[]string `json:"reasons"`
}

func (gt *GoodTime) UnmarshalJSON(b []byte) error {
	var alt altGoodTime
	alt.Reasons = &gt.Reasons
	if err := json.Unmarshal(b, &alt); err != nil {
		return err
	}
	gt.Time = time.Unix(alt.UnixTime, 0)
	return nil
}
