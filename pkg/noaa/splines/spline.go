// Package splines finds a continuous curve of tide from single points.
package splines

import (
	"bytes"
	"fmt"
	"math"
	"time"

	"github.com/spencer-p/surfdash/pkg/noaa"
)

// Curve represents a curve that links a tide event to another smoothly. Its
// derivitative at Start and End are zero and it is undefined outside Start and
// End.
type Curve struct {
	Start, End time.Time
	a, b, c, d float64
}

// A Spline is a slice of curves linked together to form a full picture.
type Spline []Curve

// CurvesBetween identifies curves to link NOAA tide predictions.
func CurvesBetween(preds noaa.Predictions) Spline {
	if len(preds) < 2 {
		return nil
	}

	curves := make([]Curve, len(preds)-1)
	for i := 0; i < len(preds)-1; i++ {
		curves[i] = curveBetween(
			time.Time(preds[i].Time),
			float64(preds[i].Height),
			time.Time(preds[i+1].Time),
			float64(preds[i+1].Height))
	}
	return curves
}

// Discrete finds n tide predictions within the tide predictions described by a
// Spline.
func Discrete(spline Spline, n int) []float64 {
	if len(spline) < 1 {
		return nil
	}
	start := []Curve(spline)[0].Start
	end := []Curve(spline)[len(spline)-1].End
	dur := end.Sub(start)
	step := time.Duration(float64(dur) / float64(n-1))

	result := make([]float64, n)
	for i := range result {
		result[i] = spline.Eval(start.Add(step * time.Duration(i)))
	}
	return result
}

func curveBetween(time1 time.Time, h1 float64, time2 time.Time, h2 float64) Curve {
	t1 := 0.0
	t2 := xrel(time1, time2)
	denominator := math.Pow(t1-t2, 3.0)
	a := (-2 * (h1 - h2)) / denominator
	b := (3 * (h1 - h2) * (t1 + t2)) / denominator
	c := (-6 * (h1 - h2) * t1 * t2) / denominator
	d := -1 * (-1*h2*math.Pow(t1, 3) + 3*h2*math.Pow(t1, 2)*t2 - 3*h1*t1*math.Pow(t2, 2) + h1*math.Pow(t2, 3)) / denominator
	curve := Curve{
		Start: time1,
		End:   time2,
		a:     a,
		b:     b,
		c:     c,
		d:     d,
	}
	return curve
}

func (s Spline) Eval(t time.Time) float64 {
	n := len(s)
	left, right := 0, n
	for right > left {
		mid := left + (right-left)/2
		if t.Before(s[mid].Start) {
			right = mid
		} else if t.After(s[mid].End) {
			left = mid
		} else {
			return s[mid].Eval(t)
		}
	}
	// Function not defined.
	return math.NaN()
}

func (c Curve) Eval(t time.Time) float64 {
	if t.Before(c.Start) || t.After(c.End) {
		return math.NaN()
	}
	x := xrel(c.Start, t)
	return c.a*x*x*x + c.b*x*x + c.c*x + c.d
}

// xrel computes an x coordinate for t that is relative to origin.
// This reduces large floating point errors by moving x coordinates closer to
// the "origin" (just the start of a particular curve).
func xrel(origin time.Time, t time.Time) float64 {
	return float64(t.Unix() - origin.Unix())
}

func (c Curve) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	_, err := fmt.Fprintf(&buf, `{"start":%d,"end":%d,"a":%g,"b":%g,"c":%g,"d":%g}`,
		c.Start.Unix(), c.End.Unix(),
		c.a, c.b, c.c, c.d)
	return buf.Bytes(), err
}
