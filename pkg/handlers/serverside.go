package handlers

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/spencer-p/surfdash/pkg/meta"
	"github.com/spencer-p/surfdash/pkg/noaa"
	"github.com/spencer-p/surfdash/pkg/sunset"
	"github.com/spencer-p/surfdash/pkg/timetricks"
	"github.com/spencer-p/surfdash/pkg/visualize"
)

const (
	minTideCookieName = "minTide"
	maxTideCookieName = "maxTide"
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

type Parameters struct {
	startDate        time.Time
	MinTide, MaxTide *http.Cookie
}

// serverSideIndex serves a good times page fully rendered on the server.
func makeServerSideIndex(content embed.FS) http.HandlerFunc {
	var indexTemplate = template.Must(template.ParseFS(content, "static/index.template.html"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := extractParameters(r)
		extendCookieLifetimes(w, params)

		// Fetch tide data first.
		query := noaa.PredictionQuery{
			// Add extra padding of one day around tides to fill in gaps.
			Start:    params.startDate.Add(-1 * 24 * time.Hour),
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
		sunevents := sunset.GetSunEvents(params.startDate, query.Duration, sunset.SantaCruz)
		// Truncate the good times predictions to account for the
		// extra data data from above.
		trimIndex := lastIndexBefore(preds, timetricks.TrimClock(params.startDate.Add(forecastLength)))
		opts := goodTimeOptionsFromParameters(params)
		goodTimes := meta.GoodTimes2(meta.Conditions{preds[:trimIndex+1], sunevents}, opts)
		tideimages := visualize.NewTidal(preds, sunevents)

		presElems := goodTimesToPresentationElements(tideimages, goodTimes)

		tinput := TemplateInput{
			PresentationElements: presElems,
			NextStart:            params.startDate.Add(forecastLength).Format(time.RFC3339),
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

func extractParameters(r *http.Request) Parameters {
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

	minTide, err := r.Cookie(minTideCookieName)
	if err != nil {
		minTide = nil
	}
	maxTide, err := r.Cookie(maxTideCookieName)
	if err != nil {
		maxTide = nil
	}

	return Parameters{
		startDate: date,
		MinTide:   minTide,
		MaxTide:   maxTide,
	}
}

func goodTimeOptionsFromParameters(p Parameters) meta.Options {
	opts := meta.Options{}
	if p.MinTide != nil {
		low, err := strconv.ParseFloat(p.MinTide.Value, 64)
		if err == nil {
			opts.LowTideThresh = &low
		}
	}
	if p.MaxTide != nil {
		high, err := strconv.ParseFloat(p.MaxTide.Value, 64)
		if err == nil {
			opts.HighTideThresh = &high
		}
	}
	return opts
}

func makeConfigTideParameters(prefix string, content embed.FS) http.HandlerFunc {
	configTideTemplate := template.Must(template.ParseFS(content, "static/config_tide.template.html"))

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			tinput := extractParameters(r)
			if err := configTideTemplate.Execute(w, tinput); err != nil {
				log.Printf("Failed to write configTideTemplate: %v", err)
			}
			return
		}

		// The remainder of this function assumes method is POST.
		if err := r.ParseForm(); err != nil {
			msg := fmt.Sprintf("Failed to parse form: %v", err)
			log.Println(msg)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, msg)
			return
		}

		params := Parameters{
			MinTide: valueAsCookie(minTideCookieName, r.PostForm.Get("min_tide")),
			MaxTide: valueAsCookie(maxTideCookieName, r.PostForm.Get("max_tide")),
		}

		extendCookieLifetimes(w, params)

		// There is only one possible referring page, so it's OK to always
		// redirect to it.
		referredFrom := prefix
		http.Redirect(w, r, referredFrom, http.StatusFound)
	}
}

func valueAsCookie(name, value string) *http.Cookie {
	if value == "" {
		return nil
	}

	// TODO: This would be a good place to tell the user it is invalid.
	if _, err := strconv.ParseFloat(value, 64); err != nil {
		log.Printf("Got %s value=%q, not a float64: %v", name, value, err)
		return nil
	}

	return &http.Cookie{
		Name:  name,
		Value: value,
	}
}

func extendCookieLifetimes(w http.ResponseWriter, p Parameters) {
	processCookie := func(name string, cookie *http.Cookie) {
		if cookie == nil {
			log.Println("Deleting cookie", name)
			// Not specified, delete it.
			http.SetCookie(w, &http.Cookie{
				Name:   name,
				MaxAge: -1, // Delete now.
				Path:   "/",
			})
			return
		}
		dayInSeconds := 60 * 60 * 24
		if cookie.MaxAge < 90*dayInSeconds {
			cookie.MaxAge += 90 * dayInSeconds
		}
		if cookie.SameSite != http.SameSiteLaxMode {
			cookie.SameSite = http.SameSiteLaxMode
		}
		if cookie.Path != "/" {
			cookie.Path = "/"
		}
		http.SetCookie(w, cookie)
	}
	processCookie(minTideCookieName, p.MinTide)
	processCookie(maxTideCookieName, p.MaxTide)
}
