package noaa

import (
	"fmt"
	"testing"
	"time"
)

func TestQueryURL(t *testing.T) {
	in := PredictionQuery{
		Start:    time.Date(2020, time.January, 5, 0, 0, 0, 0, time.Local),
		Duration: 1 * time.Hour,
		Station:  SantaCruz,
	}
	want := fmt.Sprintf("https://api.tidesandcurrents.noaa.gov/api/prod/datagetter?begin_date=20200105&datum=MLLW&end_date=20200105&format=json&interval=hilo&product=predictions&station=%d&time_zone=lst_ldt&units=english", SantaCruz)
	got := in.url().String()
	if want != got {
		t.Errorf("got  %q", got)
		t.Errorf("want %q", want)
	}
}
