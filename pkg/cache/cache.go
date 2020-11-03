package cache

import (
	"time"
)

// Timed is a cache that invalidates elements on a timer basis. It is not thread
// safe.
type Timed struct {
	ttl   time.Duration // in seconds
	cache map[string]element
}

// element holds a timestamped value to save.
type element struct {
	value    []byte
	creation time.Time
}

// NewTimed creates a new Timed cache where elements will be invalidated after
// a time in cache corresponding to TTL.
func NewTimed(ttl time.Duration) *Timed {
	return &Timed{
		ttl:   ttl,
		cache: make(map[string]element),
	}
}

// Set assigns a value to a key.
func (c *Timed) Set(key string, val []byte) {
	c.set(key, val, time.Now())
}

// set performs Set's work with the wall clock factored out.
func (c *Timed) set(key string, val []byte, t time.Time) {
	c.cache[key] = element{
		value:    val,
		creation: t,
	}
}

// Get retrieves a value for a key. The value may not exist or have expired, in
// which case ok will be false.
func (c *Timed) Get(key string) (value []byte, ok bool) {
	return c.get(key, time.Now())
}

// get is like set in that the time is factored out
func (c *Timed) get(key string, t time.Time) (value []byte, ok bool) {
	// check if the element is in memory
	el, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	// in memory elements might still be invalid
	if elapsed := t.Sub(el.creation); elapsed > c.ttl {
		delete(c.cache, key)
		return nil, false
	}

	return el.value, true
}
