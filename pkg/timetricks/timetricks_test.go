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
