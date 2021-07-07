package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/spencer-p/surfdash/pkg/meta"
	"github.com/spencer-p/surfdash/pkg/noaa"
	"github.com/spencer-p/surfdash/pkg/sunset"
	"github.com/spencer-p/surfdash/pkg/timetricks"
	"github.com/spencer-p/surfdash/pkg/visualize"
)

type PresentationElement struct {
	Date      string
	GoodTime  meta.GoodTime
	TideImage template.HTML
}

var indexTemplate = template.Must(template.ParseFiles(path.Join(getDataDir(), "static", "index.template.html")))

// serverSideIndex serves a good times page fully rendered on the server.
func serverSideIndex(w http.ResponseWriter, r *http.Request) {
	// Fetch tide data first.
	query := noaa.PredictionQuery{
		// Add extra padding of one day around tides to fill in gaps.
		Start:    time.Now().Add(-1 * 24 * time.Hour),
		Duration: forecastLength + 24*time.Hour,
		Station:  noaa.SantaCruz,
	}
	preds, err := noaa.GetPredictions(&query)
	if err != nil {
		err := fmt.Errorf("failed to fetch from NOAA: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to fetch good times: %+v", err)
		log.Printf("Failed to fetch good times: %+v", err)
		return
	}

	// Compute sun events, goodtimes, and set up tide images.
	sunevents := sunset.GetSunEvents(time.Now(), query.Duration, sunset.SantaCruz)
	// Truncate the good times predictions to account for the
	// extra data data from above.
	trimIndex := lastIndexBefore(preds, timetricks.TrimClock(time.Now().Add(forecastLength)))
	goodTimes := meta.GoodTimes2(meta.Conditions{preds[:trimIndex+1], sunevents})
	tideimages := visualize.NewTidal(preds, sunevents)

	presElems := make([]PresentationElement, len(goodTimes))
	for i := range goodTimes {
		goodTimes[i].UpdatePrettyTime()
		presElems[i] = PresentationElement{
			Date:      timetricks.Day(goodTimes[i].Time),
			GoodTime:  goodTimes[i],
			TideImage: template.HTML(imgToString(tideimages, goodTimes[i].Time)),
		}
	}

	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if err := indexTemplate.Execute(w, presElems); err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}

func imgToString(img *visualize.Tidal, t time.Time) string {
	img.SetDate(t)
	var b bytes.Buffer
	img.Encode(&b)
	return b.String()
}

// TODO(spencer-p) Standardize these scattered binary search functions.
func lastIndexBefore(preds noaa.Predictions, t time.Time) int {
	left, right := 0, len(preds)
	for right-left > 1 {
		mid := (left + right) / 2
		midt := preds[mid].T()
		if midt.Before(t) {
			left = mid
		} else if midt.After(t) {
			right = mid
		} else if midt.Equal(t) {
			return mid
		}
	}
	ok := left < len(preds)
	if !ok {
		// Nothing found, just return last element
		return len(preds) - 1
	}
	return left
}
