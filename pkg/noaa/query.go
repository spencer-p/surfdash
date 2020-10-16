package noaa

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	NOAA_URL = "https://api.tidesandcurrents.noaa.gov/api/prod/datagetter"
	TIME_FMT = "20060102"
)

func GetPredictions(q *PredictionQuery) (PredictionList, error) {
	var result Predictions

	// Build request URL first
	addr, err := url.Parse(NOAA_URL)
	if err != nil {
		return nil, err
	}

	addr.RawQuery = q.build().Encode()

	// Make the request to NOAA
	resp, err := http.Get(addr.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Predictions, nil
}

func (q *PredictionQuery) build() url.Values {
	vals := make(url.Values)
	vals.Add("begin_date", q.Start.Format(TIME_FMT))
	vals.Add("end_date", q.End.Format(TIME_FMT))
	vals.Add("station", fmt.Sprintf("%d", q.Station))
	vals.Add("product", "predictions")
	vals.Add("datum", "MLLW")
	vals.Add("time_zone", "lst_ldt")
	vals.Add("interval", "hilo")
	vals.Add("units", "english")
	vals.Add("format", "json")
	return vals
}
