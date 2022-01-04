package metrics

import (
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "request_latency",
			Subsystem: "surfdash",
			Help:      "HTTP request latencies in seconds.",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.2, 0.4, 0.8, 1.0, 2.0, 4.0, 8.0, 16.0, 32.0},
		},
		[]string{"verb", "path", "code"},
	)
)

func init() {
	prometheus.MustRegister(
		requestLatency,
	)
}

func ObserveRequestLatency(verb, path, code string, latency float64) {
	requestLatency.With(prometheus.Labels{
		"code": code,
		"verb": verb,
		"path": path,
	}).Observe(latency)
}

func LatencyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		verb := r.Method
		path := ""
		if r.URL != nil {
			path = r.URL.Path
		}

		// Defer metric observing. Any panics in next are reported as 500 errors
		// and then re-thrown.
		defer func() {
			if err := recover(); err != nil {
				ObserveRequestLatency(verb, path, "500", time.Now().Sub(t).Seconds())
				panic(err)
			}
			code := getStatusCode(w)
			ObserveRequestLatency(verb, path, code, time.Now().Sub(t).Seconds())
		}()

		next.ServeHTTP(w, r)
	})
}

func getStatusCode(w http.ResponseWriter) string {
	statusFields, ok := w.Header()["Status-Code"]
	if !ok {
		// Unset, will be set to 200 by stdlib.
		return "200"
	}
	if len(statusFields) < 1 {
		// Not normal behavior.
		return "0"
	}
	return statusFields[0]
}
