package sunset

import (
	"fmt"
	"time"
)

// Place is a lat/long coordinate on the Earth matched with its time zone.
type Place struct {
	Lat, Long float64
	Location  *time.Location
}

var (
	SantaCruz = Place{
		36.9741, -122.0308,
		locationOrPanic("America/Los_Angeles"),
	}
)

// SunEventList is a time series of SunEvent.
type SunEventList []SunEvent

// SunEvent is a sunrise or sunset event.
type SunEvent struct {
	Time  time.Time
	Event Event
}

func (s *SunEvent) String() string {
	return fmt.Sprintf("%s %s",
		s.Time.Format(time.RFC822),
		func() string {
			if s.Event == Sunrise {
				return "Sunrise"
			} else {
				return "Sunset"
			}
		}())
}

// Event encodes a sunrise or sunset event.
type Event bool

const (
	Sunrise Event = true
	Sunset        = false
)

func locationOrPanic(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}
