package cache

import (
	"testing"
	"time"
)

func TestTimed(t *testing.T) {
	c := NewTimed(5 * time.Minute)

	tstart := time.Now()

	c.set("key", []byte("value"), tstart)

	_, ok := c.get("key", tstart.Add(time.Minute))
	if !ok {
		t.Errorf("failed to get key that shoul not be expired")
	}

	_, ok = c.get("key", tstart.Add(10*time.Minute))
	if ok {
		t.Errorf("succeeded in getting expired key")
	}

	_, ok = c.get("key", tstart.Add(time.Minute))
	if ok {
		t.Errorf("succeeded in getting key that was previously evicted")
	}
}
