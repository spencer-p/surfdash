package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/spencer-p/surfdash/pkg/cache"
	"github.com/spencer-p/surfdash/pkg/meta"
	"github.com/spencer-p/surfdash/pkg/noaa"
	"github.com/spencer-p/surfdash/pkg/sunset"

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
	r.Handle("/api/v1/goodtimes", makeServeGoodTimes())
	r.PathPrefix("/static/").Handler(http.StripPrefix(prefix, http.FileServer(http.Dir(dataDir))))
}

func getDataDir() string {
	if dir := os.Getenv(koDataEnvKey); dir != "" {
		return dir
	} else {
		return "."
	}
}

func fetchGoodTimes(numDays time.Duration) ([]meta.GoodTime, error) {
	query := noaa.PredictionQuery{
		Start:    time.Now(),
		Duration: numDays * day,
		Station:  noaa.SantaCruz,
	}

	preds, err := noaa.GetPredictions(&query)
	if err != nil {
		return nil, err
	}

	sunevents := sunset.GetSunEvents(time.Now(), query.Duration, sunset.SantaCruz)

	goodTimes := meta.GoodTimes(meta.Conditions{preds, sunevents})
	return goodTimes, nil
}

func makeServeGoodTimes() http.Handler {
	// cache for slightly less than one day so daily clients don't see stale
	// data
	timeCache := cache.NewTimed(23 * time.Hour)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// cache based on method and URL, which should encapsulate the query
		key := fmt.Sprintf("%s %s", r.Method, r.URL)

		// serve cache version from memory if possible
		if cached, ok := timeCache.Get(key); ok {
			w.Header().Add("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write(cached)
			return
		}
		log.Println("No cache data")

		goodTimes, err := fetchGoodTimes(forecastLength)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Failed to get data: %+v", err)
			log.Printf("Failed to get data: %+v", err)
			return
		}

		// duplicate the http response onto a buffer for the cache
		var toCache bytes.Buffer
		mw := io.MultiWriter(w, &toCache)

		// serve result
		outputFormat := r.FormValue("o")
		if outputFormat == "json" {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(mw).Encode(goodTimes); err != nil {
				log.Printf("Failed to encode JSON result: %+v", err)
			}
		} else {
			w.Header().Add("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			for i, gt := range goodTimes {
				fmt.Fprintf(mw, "%s", gt.String())
				if i+1 < len(goodTimes) {
					fmt.Fprintf(mw, "\n")
				}
			}
		}

		// save the result asynchonously as the cache may block
		go func() {
			timeCache.Set(key, toCache.Bytes())
		}()
	})
}

func makeIndexHandler() http.Handler {
	file := path.Join(getDataDir(), "static", "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, file)
	})
}
