package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"

	"github.com/spencer-p/surfdash/pkg/meta"
	"github.com/spencer-p/surfdash/pkg/noaa"
	"github.com/spencer-p/surfdash/pkg/sunset"
)

type Config struct {
	Port   string `default:"8080"`
	Prefix string `default:"/"`
}

func fetchGoodTimes(numDays int) ([]meta.GoodTime, error) {
	query := noaa.PredictionQuery{
		Start:    time.Now(),
		Duration: time.Duration(numDays) * 24 * time.Hour,
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

func serveGoodTimes(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	w.Header().Add("Content-Type", "text/plain")
	goodTimes, err := fetchGoodTimes(7) // 7 days of forecast
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to get data: %+v", err)
		log.Printf("Failed to get data: %+v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	for _, gt := range goodTimes {
		fmt.Fprintf(w, "%s\n", gt.String())
	}
}

func main() {
	var env Config
	if err := envconfig.Process("", &env); err != nil {
		log.Fatal(err.Error())
	}

	r := mux.NewRouter().StrictSlash(true)
	s := r.PathPrefix(env.Prefix).Subrouter()

	s.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		fmt.Fprintf(w, "hello world\n")
	})
	s.HandleFunc("/api/v1/goodtimes", serveGoodTimes)

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:" + env.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Listening and serving on %s/%s", srv.Addr, env.Prefix[1:])
	log.Fatal(srv.ListenAndServe())
}
