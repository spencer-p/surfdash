package handlers

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/spencer-p/surfdash/pkg/meta"
	"github.com/spencer-p/surfdash/pkg/noaa"
	"github.com/spencer-p/surfdash/pkg/sunset"
	"github.com/spencer-p/surfdash/pkg/timetricks"
	"github.com/spencer-p/surfdash/pkg/visualize"
)

type TemplateInput struct {
	PresentationElements []PresentationElement
	NextStart            string
}

type PresentationElement struct {
	Date      string
	GoodTimes []meta.GoodTime
	TideImage template.HTML
}

// serverSideIndex serves a good times page fully rendered on the server.
func makeServerSideIndex(content embed.FS) http.HandlerFunc {
	var indexTemplate = template.Must(template.ParseFS(content, "static/index.template.html"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		date := time.Now()
		startString := r.FormValue("start")
		if startString != "" {
			parsed, err := time.Parse(time.RFC3339, startString)
			if err != nil {
				log.Printf("Failed to read time %q: %v", startString, err)
			} else {
				date = parsed
			}
		}

		// Fetch tide data first.
		query := noaa.PredictionQuery{
			// Add extra padding of one day around tides to fill in gaps.
			Start:    date.Add(-1 * 24 * time.Hour),
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
		sunevents := sunset.GetSunEvents(date, query.Duration, sunset.SantaCruz)
		// Truncate the good times predictions to account for the
		// extra data data from above.
		trimIndex := lastIndexBefore(preds, timetricks.TrimClock(date.Add(forecastLength)))
		goodTimes := meta.GoodTimes2(meta.Conditions{preds[:trimIndex+1], sunevents})
		tideimages := visualize.NewTidal(preds, sunevents)

		presElems := goodTimesToPresentationElements(tideimages, goodTimes)

		tinput := TemplateInput{
			PresentationElements: presElems,
			NextStart:            date.Add(forecastLength).Format(time.RFC3339),
		}

		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		if err := indexTemplate.Execute(w, tinput); err != nil {
			log.Printf("Failed to execute template: %v", err)
		}
	})
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

func goodTimesToPresentationElements(tideimages *visualize.Tidal, goodTimes []meta.GoodTime) []PresentationElement {
	var f func(result []PresentationElement, goodTimes []meta.GoodTime) []PresentationElement
	f = func(result []PresentationElement, goodTimes []meta.GoodTime) []PresentationElement {
		if len(goodTimes) == 0 {
			return result
		}

		resultLen := len(result)
		gt := goodTimes[0]
		gt.UpdatePrettyTime()

		if len(result) != 0 && result[resultLen-1].Date == timetricks.Day(gt.Time) {
			// There is already an entry in the result that corresponds to the
			// same day as the next time we're entering.
			result[resultLen-1].GoodTimes = append(result[resultLen-1].GoodTimes, gt)
		} else {
			// Normal case.
			result = append(result, PresentationElement{
				Date:      timetricks.Day(gt.Time),
				GoodTimes: []meta.GoodTime{gt},
				TideImage: template.HTML(imgToString(tideimages, gt.Time)),
			})
		}

		return f(result, goodTimes[1:])
	}

	return f(nil, goodTimes)
}
