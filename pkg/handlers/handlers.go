package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/spencer-p/surfdash/pkg/cache"
	"github.com/spencer-p/surfdash/pkg/meta"
	"github.com/spencer-p/surfdash/pkg/noaa"
	"github.com/spencer-p/surfdash/pkg/sunset"
	"github.com/spencer-p/surfdash/pkg/timetricks"
	"github.com/spencer-p/surfdash/pkg/visualize"

	"github.com/gorilla/mux"
)

const (
	day            = 24 * time.Hour
	forecastLength = 7 * day
	cacheTTL       = 1 * day

	koDataEnvKey = "KO_DATA_PATH"
)

func Register(r *mux.Router, prefix string) {
	dataDir := getDataDir()

	r.Handle("/", makeIndexHandler())
	r.HandleFunc("/api/v1/goodtimes", serveGoodTimes)
	r.HandleFunc("/api/v2/goodtimes", serveGoodTimes2)
	r.HandleFunc("/api/v2/tide_image", serveTideImage)
	r.PathPrefix("/static/").Handler(http.StripPrefix(prefix, http.FileServer(http.Dir(dataDir))))
}

func getDataDir() string {
	if dir := os.Getenv(koDataEnvKey); dir != "" {
		return dir
	} else {
		return "."
	}
}

func makeFetchGoodTimes() func(time.Duration) ([]meta.GoodTime, error) {
	// cache for an hour at a time.
	timeCache := cache.NewTimed(1 * time.Hour)

	return func(dur time.Duration) ([]meta.GoodTime, error) {
		// serve cache version from memory if possible
		key := timetricks.UniqueDay(time.Now()) + dur.String()
		if cached, ok := timeCache.Get(key); ok {
			var goodTimes []meta.GoodTime
			if err := json.Unmarshal(cached, &goodTimes); err != nil {
				return nil, fmt.Errorf("failed to unmarshal from cache: %w", err)
			}
			return goodTimes, nil
		}
		log.Println("No cache data")

		query := noaa.PredictionQuery{
			Start:    time.Now(),
			Duration: dur,
			Station:  noaa.SantaCruz,
		}

		preds, err := noaa.GetPredictions(&query)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch from NOAA: %w", err)
		}

		sunevents := sunset.GetSunEvents(time.Now(), query.Duration, sunset.SantaCruz)

		goodTimes := meta.GoodTimes(meta.Conditions{preds, sunevents})

		// save the result to cache asynchonously as it may block
		go func() {
			toCache, err := json.Marshal(&goodTimes)
			if err != nil {
				log.Println("Failed to cache good times:", err)
				return // this error is lost to the user
			}
			timeCache.Set(key, toCache)
		}()

		return goodTimes, nil
	}
}

var fetchGoodTimes = makeFetchGoodTimes()

func serveGoodTimes(w http.ResponseWriter, r *http.Request) {
	// get the good times
	goodTimes, err := fetchGoodTimes(forecastLength)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to fetch good times: %+v", err)
		log.Printf("Failed to fetch good times: %+v", err)
		return
	}

	// serve result
	outputFormat := r.FormValue("o")
	if outputFormat == "json" {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(goodTimes); err != nil {
			log.Printf("Failed to encode JSON result: %+v", err)
		}
	} else {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		for i, gt := range goodTimes {
			fmt.Fprintf(w, "%s", gt.String())
			if i+1 < len(goodTimes) {
				fmt.Fprintf(w, "\n")
			}
		}
		if len(goodTimes) == 0 {
			fmt.Fprintf(w, "No good times found.")
		}
	}
}

func makeIndexHandler() http.Handler {
	file := path.Join(getDataDir(), "static", "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, file)
	})
}

func serveGoodTimes2(w http.ResponseWriter, r *http.Request) {
	// get the good times
	goodTimes, err := fetchGoodTimes2(forecastLength)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to fetch good times: %+v", err)
		log.Printf("Failed to fetch good times: %+v", err)
		return
	}

	// serve result
	outputFormat := r.FormValue("o")
	if outputFormat == "json" {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(goodTimes); err != nil {
			log.Printf("Failed to encode JSON result: %+v", err)
		}
	} else {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		for i, gt := range goodTimes {
			fmt.Fprintf(w, "%s", gt.String())
			if i+1 < len(goodTimes) {
				fmt.Fprintf(w, "\n")
			}
		}
		if len(goodTimes) == 0 {
			fmt.Fprintf(w, "No good times found.")
		}
	}
}

func fetchGoodTimes2(dur time.Duration) ([]meta.GoodTime, error) {
	query := noaa.PredictionQuery{
		Start:    time.Now(),
		Duration: dur,
		Station:  noaa.SantaCruz,
	}

	preds, err := noaa.GetPredictions(&query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from NOAA: %w", err)
	}

	sunevents := sunset.GetSunEvents(time.Now(), query.Duration, sunset.SantaCruz)

	goodTimes := meta.GoodTimes2(meta.Conditions{preds, sunevents})

	return goodTimes, nil
}

func serveTideImage(w http.ResponseWriter, r *http.Request) {
	query := noaa.PredictionQuery{
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

	sunevents := sunset.GetSunEvents(time.Now(), query.Duration, sunset.SantaCruz)

	date := r.FormValue("t")
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		log.Printf("Failed to read time %q: %v", date, err)
		t = time.Now()
	}
	img := visualize.NewTidal(preds, sunevents)
	img.SetDate(t)
	w.Header().Add("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	img.Encode(w)
}
