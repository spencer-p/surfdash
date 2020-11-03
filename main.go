package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/spencer-p/helpttp"
)

type Config struct {
	Port   string `default:"8080"`
	Prefix string `default:"/"`
}

func main() {
	var env Config
	if err := envconfig.Process("", &env); err != nil {
		log.Fatal(err.Error())
	}

	r := mux.NewRouter().StrictSlash(true)
	r.Use(helpttp.WithLog)
	s := r.PathPrefix(env.Prefix).Subrouter()

	s.HandleFunc("/", handleIndex)
	s.Handle("/api/v1/goodtimes", makeServeGoodTimes())

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:" + env.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Listening and serving on %s/%s", srv.Addr, env.Prefix[1:])
	log.Fatal(srv.ListenAndServe())
}
