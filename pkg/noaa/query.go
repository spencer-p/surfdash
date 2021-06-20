package noaa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/spencer-p/surfdash/pkg/cache"
)

const QUERY_TIME_FMT = "20060102"

var NOAA_URL = url.URL{
	Scheme: "https",
	Host:   "api.tidesandcurrents.noaa.gov",
	Path:   "/api/prod/datagetter",
}

var qcache = cache.NewTimed(12 * time.Hour)

// GetPredictions builds a query and sends a request to NOAA for tide prediction
// data.
func GetPredictions(q *PredictionQuery) (Predictions, error) {
	// Build request URL first
	addr := q.url().String()

	var body []byte
	var inCache bool
	body, inCache = qcache.Get(addr)
	if !inCache {
		// Make the request to NOAA
		resp, err := http.Get(addr)
		if err != nil {
			return nil, fmt.Errorf("failed GET request: %w", err)
		}
		defer resp.Body.Close()

		// Read the full response to a buffer for parsing and caching.
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, resp.Body); err != nil {
			return Predictions{}, fmt.Errorf("failed to read NOAA response: %w", err)
		}
		body = buf.Bytes()
	}

	reader := bytes.NewReader(body)
	result, err := decodeResponse(reader)
	if err != nil {
		return Predictions{}, fmt.Errorf("failed to parse NOAA response: %w", err)
	}

	if !inCache {
		qcache.Set(addr, body)
	}
	return result.Predictions, nil
}

func (q *PredictionQuery) url() *url.URL {
	addr := NOAA_URL
	addr.RawQuery = q.build().Encode()
	return &addr
}

func (q *PredictionQuery) build() url.Values {
	vals := make(url.Values)
	vals.Add("begin_date", q.Start.Format(QUERY_TIME_FMT))
	vals.Add("end_date", q.Start.Add(q.Duration).Format(QUERY_TIME_FMT))
	vals.Add("station", fmt.Sprintf("%d", q.Station))
	vals.Add("product", "predictions")
	vals.Add("datum", "MLLW")
	vals.Add("time_zone", "lst_ldt")
	vals.Add("interval", "hilo")
	vals.Add("units", "english")
	vals.Add("format", "json")
	return vals
}

func decodeResponse(resp io.Reader) (*NOAAResult, error) {
	var result NOAAResult
	if err := json.NewDecoder(resp).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
