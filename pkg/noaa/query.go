package noaa

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const QUERY_TIME_FMT = "20060102"

var NOAA_URL = url.URL{
	Scheme: "https",
	Host:   "api.tidesandcurrents.noaa.gov",
	Path:   "/api/prod/datagetter",
}

// GetPredictions builds a query and sends a request to NOAA for tide prediction
// data.
func GetPredictions(q *PredictionQuery) (Predictions, error) {
	// Build request URL first
	addr := q.url()

	// Make the request to NOAA
	resp, err := http.Get(addr.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, err := decodeResponse(resp.Body)
	if err != nil {
		return Predictions{}, err
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
