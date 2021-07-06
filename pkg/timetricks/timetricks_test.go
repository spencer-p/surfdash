package timetricks

import (
	"fmt"
	"time"
)

func ExampleWithinWeek() {
	t := time.Now()
	for i := 0; i < 8; i++ {
		fmt.Println(i, WithinWeek(t.Add(time.Duration(i)*24*time.Hour)))
	}
	// Output:
	// 0 true
	// 1 true
	// 2 true
	// 3 true
	// 4 true
	// 5 true
	// 6 true
	// 7 false
}

func ExampleTrimClock() {
	t := time.Date(2020, 03, 14, 19, 45, 6, 500, time.UTC)
	midnight := TrimClock(t)
	fmt.Println(midnight.String())
	// Output:
	// 2020-03-14 00:00:00 +0000 UTC
}
