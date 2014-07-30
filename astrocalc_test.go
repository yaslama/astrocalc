package astrocalc

import (
	"math"
	"strings"
	"testing"
	"time"
)

func testNear(t *testing.T, subject string, current, good float64) {
	if math.Abs(current-good) >= 1E-15 {
		t.Errorf("%s: %.20f instead of %.20f\n", subject, current, good)
	}
}

func TestNewJulian1(t *testing.T) {
	t1, _ := time.Parse("Jan 2 2006 15:04:05", "Apr 1 2004 12:00:00")
	j1 := NewJulian(t1)
	if j1.julianDayNumber != 2453097 || j1.time != 0 {
		t.Errorf("Error in %#v\n", j1)
	}
}

func TestNewJulian2(t *testing.T) {
	t1, _ := time.Parse("Jan 2 2006 15:04:05", "Jul 29 2014 19:03:25")
	j1 := NewJulian(t1)
	if j1.julianDayNumber != 2456868 || j1.time != 25405*1e9 {
		t.Errorf("Error in %#v\n", j1)
	}
}

func TestSunPosition(t *testing.T) {
	t1, _ := time.Parse("Jan 2 2006 15:04:05", "Jul 29 2014 19:03:25")
	lat := 31.783
	lng := 35.233
	sunCalc := NewSunCalc()
	azim, alti := sunCalc.GetPosition(t1, lat, lng)
	testNear(t, "azimuth", azim, 2.3820139121247865)
	testNear(t, "altitude", alti, -0.4573946150014954)
}

func TestTimes(t *testing.T) {
	t1, _ := time.Parse("Jan 2 2006 15:04:05", "Jul 29 2014 19:03:25")
	lat := 31.783
	lng := 35.233
	timesGood := map[string]string{
		"goldenHour":    "2014-07-29T16:05:16.619633138Z",
		"dawn":          "2014-07-29T02:27:04.727511405Z",
		"nauticalDusk":  "2014-07-29T17:38:37.470324039Z",
		"nightEnd":      "2014-07-29T01:20:46.21797055Z",
		"night":         "2014-07-29T18:12:41.606369912Z",
		"solarNoon":     "2014-07-29T09:46:43.912170231Z",
		"dusk":          "2014-07-29T17:06:23.096829056Z",
		"sunsetStart":   "2014-07-29T16:36:54.901068806Z",
		"nauticalDawn":  "2014-07-29T01:54:50.354016423Z",
		"sunset":        "2014-07-29T16:39:37.948705852Z",
		"sunriseEnd":    "2014-07-29T02:56:32.923271656Z",
		"goldenHourEnd": "2014-07-29T03:28:11.204707324Z",
		"nadir":         "2014-07-28T21:46:43.912170231Z",
		"sunrise":       "2014-07-29T02:53:49.87563461Z",
	}
	sunCalc := NewSunCalc()
	times := sunCalc.GetTimes(t1, lat, lng)
	for name, t2 := range times {
		t2UTC := t2.UTC().Format(time.RFC3339Nano)
		if strings.TrimSpace(t2UTC) != strings.TrimSpace(timesGood[name]) {
			t.Errorf("%s: %s instead of %s", name, t2UTC, timesGood[name])
		}
	}
}

func TestGetMoonPosition(t *testing.T) {
	t1, _ := time.Parse("Jan 2 2006 15:04:05", "Jul 29 2014 19:03:25")
	lat := 31.783
	lng := 35.233
	azim, alti, dist := GetMoonPosition(t1, lat, lng)
	testNear(t, "azimuth", azim, 1.8424006017910686)
	testNear(t, "altitude", alti, -0.2419311867071057)
	testNear(t, "distance", dist, 404133.76960804936)
}

func TestGetMoonIllumination(t *testing.T) {
	t1, _ := time.Parse("Jan 2 2006 15:04:05", "Jul 29 2014 19:03:25")
	fraction, phase, angle := GetMoonIllumination(t1)
	testNear(t, "fraction", fraction, 0.07382281607579783)
	testNear(t, "phase", phase, 0.08758701583098588)
	testNear(t, "angle", angle, -1.0922384803528917)
}
