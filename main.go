package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/spencer-p/surfdash/pkg/handlers"
	"github.com/spencer-p/surfdash/pkg/metrics"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spencer-p/helpttp"
)

//go:embed static
var staticContent embed.FS

type Config struct {
	Port        string `default:"8080"`
	MetricsPort string `default:"8081"`
	Prefix      string `default:"/"`
}

func main() {
	var env Config
	if err := envconfig.Process("", &env); err != nil {
		log.Fatal(err.Error())
	}

	r := mux.NewRouter().StrictSlash(true)
	r.Use(helpttp.WithLog)
	r.Use(metrics.LatencyHandler)
	s := r.PathPrefix(env.Prefix).Subrouter()
	handlers.Register(s, env.Prefix, staticContent)

	if env.Prefix != "/" {
		r.Handle("/", http.RedirectHandler(env.Prefix, http.StatusFound))
	}

	metricsSrv := &http.Server{
		Handler:      promhttp.Handler(),
		Addr:         "0.0.0.0:" + env.MetricsPort,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go func() {
		log.Printf("Listening and serving metrics on %s", metricsSrv.Addr)
		log.Fatal(metricsSrv.ListenAndServe())
	}()

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:" + env.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Listening and serving on %s/%s", srv.Addr, env.Prefix[1:])
	log.Fatal(srv.ListenAndServe())
}
