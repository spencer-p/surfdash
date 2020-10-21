package noaa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestParsePrediction(t *testing.T) {
	table := []struct {
		input string
		want  Prediction
	}{{
		input: `{"t":"2020-10-20 02:17", "v":"4.080", "type":"H"}`,
		want: Prediction{
			Time:   Time(time.Date(2020, time.October, 20, 2, 17, 0, 0, time.Local)),
			Height: 4.08,
			Type:   HighTide,
		},
	}, {
		input: `{"t":"2019-09-21 06:56", "v":"2.559", "type":"L"}`,
		want: Prediction{
			Time:   Time(time.Date(2019, time.September, 21, 6, 56, 0, 0, time.Local)),
			Height: 2.559,
			Type:   LowTide,
		},
	}}

	for _, test := range table {
		t.Run(test.input, func(t *testing.T) {
			var got Prediction

			dec := json.NewDecoder(bytes.NewBufferString(test.input))
			if err := dec.Decode(&got); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}

			gotstr := fmt.Sprintf("%s", got)
			wantstr := fmt.Sprintf("%s", test.want)
			if diff := cmp.Diff(gotstr, wantstr); diff != "" {
				t.Errorf("incorrect parse (-got,+want): %s", diff)
			}
		})
	}
}
