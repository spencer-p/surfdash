// Package noaa implements queries to NOAA to retrieve tide data.  Tide data is
// requested as a time series per location (see PredictionQuery).  A successful
// query returns a list of predictions with time, height, and whether it is high
// or low. All times are local.
package noaa
