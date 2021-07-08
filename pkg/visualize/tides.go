package visualize

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/spencer-p/surfdash/pkg/noaa"
	"github.com/spencer-p/surfdash/pkg/noaa/splines"
	"github.com/spencer-p/surfdash/pkg/sunset"
	"github.com/spencer-p/surfdash/pkg/timetricks"
)

const (
	width  = 1200
	height = 300
)

type Tidal struct {
	date      time.Time
	tidePreds noaa.Predictions
	sunEvents sunset.SunEvents
}

func NewTidal(tidePreds noaa.Predictions, sunEvents sunset.SunEvents) *Tidal {
	return &Tidal{
		tidePreds: tidePreds,
		sunEvents: sunEvents,
	}
}

func (img *Tidal) SetDate(t time.Time) {
	img.date = timetricks.TrimClock(t)
}

func (img *Tidal) Encode(w io.Writer) (int, error) {
	var n int
	var err error
	io := func(nextn int, nexterr error) {
		n += nextn
		if nexterr != nil {
			err = nexterr
		}
	}

	io(fmt.Fprintf(w, `<svg viewBox="0 0 %d %d" onclick="" xmlns="http://www.w3.org/2000/svg">`, width, height))

	// Calculate dawn/dusk and draw the sunshine.
	sunupIndex, ok := img.sunup(img.date)
	if !ok || sunupIndex+1 > len(img.sunEvents) {
		return n, fmt.Errorf("Not enough sun data")
	}
	sunup := img.sunEvents[sunupIndex]
	sundown := img.sunEvents[sunupIndex+1]
	risex := img.timeToX(sunup.Time)
	setx := img.timeToX(sundown.Time)
	io(fmt.Fprintf(w, `<rect class="daytime" fill="lightyellow" x="%d" y="%d" width="%d" height="%d"/>`,
		risex, 0,
		setx-risex, height))

	// Draw markers for tide levels.
	io(fmt.Fprintf(w, `<rect class="two_foot" fill="#e76f51" x="%d" y="%d" width="%d" height="%d"/>`,
		0, tideHeightToY(2),
		width, tideHeightToY(1)-tideHeightToY(2)+1))
	io(fmt.Fprintf(w, `<rect class="one_foot" fill="#f4a261" x="%d" y="%d" width="%d" height="%d"/>`,
		0, tideHeightToY(1),
		width, tideHeightToY(0)-tideHeightToY(1)+1))
	io(fmt.Fprintf(w, `<rect class="zero_foot" fill="#e9c46a" x="%d" y="%d" width="%d" height="%d"/>`,
		0, tideHeightToY(0),
		width, tideHeightToY(-2)-tideHeightToY(0)+1))

	// Choose the first tide prediction to start from. Should be off screen; if
	// not, just start at the beginning.
	i, ok := img.indexPredPreceding(img.date)
	if !ok {
		i = 0
	}
	startPredI, endPredI := i, i

	for ; i+1 < len(img.tidePreds); i += 1 {
		x1 := img.timeToX(img.tidePreds[i].T())
		y1 := tideHeightToY(img.tidePreds[i].Height)
		if int(x1) > width {
			break
		}
		endPredI = i + 1
		io(fmt.Fprintf(w, `<path class="tide" fill="skyblue" d="M %d,%d `, x1, y1))

		x2 := img.timeToX(img.tidePreds[i+1].T()) + 1 // +1 to create overlap
		y2 := tideHeightToY(img.tidePreds[i+1].Height)

		cx1, cy1 := (x1+x2)/2, y1
		cx2, cy2 := cx1, y2

		io(fmt.Fprintf(w, `C %d,%d %d,%d %d,%d `,
			cx1, cy1,
			cx2, cy2,
			x2, y2))

		io(fmt.Fprintf(w, `L %d,%d L %d,%d z"/>`, x2, height, x1, height))
	}

	// Draw the night time shadows.
	io(fmt.Fprintf(w, `<rect class="night" fill="blue" fill-opacity="25%%" x="%d" y="%d" width="%d" height="%d"/>`,
		0, 0,
		risex, height))
	io(fmt.Fprintf(w, `<rect class="night" fill="blue" fill-opacity="25%%" x="%d" y="%d" width="%d" height="%d"/>`,
		setx, 0,
		width-setx, height))

	// Insert spline data as JSON.
	splinePreds := img.tidePreds[startPredI : endPredI+1]
	spline := splines.CurvesBetween(splinePreds)
	io(fmt.Fprintf(w, `<text class="spline" visibility="hidden">`))
	json.NewEncoder(w).Encode(spline)
	io(fmt.Fprintf(w, `</text>`))

	// Insert date of this graph as unix.
	io(fmt.Fprintf(w, `<text class="unixtime" visibility="hidden">%d</text>`, img.date.Unix()))

	io(fmt.Fprintf(w, `</svg>`))

	return n, err
}

func (img *Tidal) indexPredPreceding(t time.Time) (int, bool) {
	left, right := 0, len(img.tidePreds)
	for right-left > 1 {
		mid := (left + right) / 2
		midt := img.tidePreds[mid].T()
		if midt.Before(t) {
			left = mid
		} else if midt.After(t) {
			right = mid
		} else if midt.Equal(t) {
			return mid, true
		}
	}
	ok := left < len(img.tidePreds)
	return left, ok
}

func (img *Tidal) sunup(t time.Time) (int, bool) {
	for i := 0; i < len(img.sunEvents); i++ {
		if img.sunEvents[i].Time.After(t) {
			return i, true
		}
	}
	return 0, false
}

func tideHeightToY(tideHeight noaa.Height) int {
	return height - int((tideHeight+2)*(height/10)) // scaling ratio of img height to 10 feet of tide variance
}

func (img *Tidal) timeToX(t time.Time) int {
	return int(t.Unix()-img.date.Unix()) * width / (60 * 60 * 24)
}
