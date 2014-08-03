package astrocalc

import (
	"math"
	"time"
)

// A JulianDate represents a Julian date as defined
// at http://aa.quae.nl/en/reken/juliaansedag.html
type JulianDate struct {
	julianDayNumber int64
	time            int64 //nanoseconds since the beginning of the day
}

func integerDivide(a, b int64) (q, r int64) {
	q = a / b
	r = a % b
	if r < 0 {
		q = q - 1
		r = r + b
	}
	return
}

// sun calculations are based on http://aa.quae.nl/en/reken/zonpositie.html formulas

// date/time constants and conversions

const (
	daySec      = 60 * 60 * 24
	halfDaySec  = 60 * 60 * 12
	halfDayNano = 1e9 * 60 * 60 * 12
	j1970       = 2440588
	j2000       = 2451545
)

func (j *JulianDate) removeHalfDay() {
	if j.time < halfDayNano {
		j.time = j.time + halfDayNano
		j.julianDayNumber = j.julianDayNumber - 1
	} else {
		j.time = j.time - halfDayNano
	}
}

// NewJulian creates a new JulianDate from a time.Time
func NewJulian(t time.Time) JulianDate {
	d1970, secs := integerDivide(t.Unix(), daySec)
	j := d1970 + j1970
	jdn := JulianDate{
		julianDayNumber: j,
		time:            secs*1e9 + int64(t.Nanosecond()),
	}
	jdn.removeHalfDay()
	return jdn
}

func toDays(t time.Time) JulianDate {
	jdn := NewJulian(t)
	jdn.julianDayNumber = jdn.julianDayNumber - j2000
	return jdn
}

// Time returns a time.Time object corresponding to the Julian Date
func (j JulianDate) Time() time.Time {
	return time.Unix((j.julianDayNumber-j1970)*daySec+halfDaySec, j.time)
}

// JulianFromFloat create a JulianDate from a float
func JulianFromFloat(j float64) JulianDate {
	int, frac := math.Modf(j)
	return JulianDate{
		julianDayNumber: int64(int),
		time:            int64(frac * daySec * 1e9),
	}
}

// DayTime returns the julianDayNumber and the nanoseconds since the
//  beginning of the day
func (j JulianDate) DayTime() (julianDayNumber, time int64) {
	return j.julianDayNumber, j.time
}

// JulianFromDayTime creates a JulianDate from a julianDaynumber and the
//  nanoseconds since the beginning of the day
func JulianFromDayTime(julianDayNumber, time int64) JulianDate {
	jdn := JulianDate{
		julianDayNumber: julianDayNumber,
		time:            time,
	}
	return jdn
}

// general calculations for position

const rad = math.Pi / 180
const (
	e           = rad * 23.4397 // obliquity of the Earth
	stCoef0Nano = 28016 * 1e7
	stCoef1Nano = 3609856235 * 100
)

func rightAscension(l, b float64) float64 {
	return math.Atan2(math.Sin(l)*math.Cos(e)-math.Tan(b)*math.Sin(e), math.Cos(l))
}
func declination(l, b float64) float64 {
	return math.Asin(math.Sin(b)*math.Cos(e) + math.Cos(b)*math.Sin(e)*math.Sin(l))
}

func azimuth(H, phi, dec float64) float64 {
	return math.Atan2(math.Sin(H), math.Cos(H)*math.Sin(phi)-math.Tan(dec)*math.Cos(phi))
}
func altitude(H, phi, dec float64) float64 {
	return math.Asin(math.Sin(phi)*math.Sin(dec) + math.Cos(phi)*math.Cos(dec)*math.Cos(H))
}

func siderealTime(d JulianDate, lw float64) float64 {
	//return rad*(280.16+360.9856235*d) - lw
	i := (int64(stCoef0Nano) + d.julianDayNumber*stCoef1Nano) % (360 * 1e9)
	st := rad*(float64(i)/1e9+float64(d.time)/daySec*(float64(stCoef1Nano)/1e18)) - lw
	if st < 0 {
		st = st + 2*math.Pi
	}
	return st
}

// general sun calculations
const (
	maCoef0Nano = 3575291 * 1e5
	maCoef1Nano = 985600280
)

func solarMeanAnomaly(d JulianDate) float64 {
	//return rad * (357.5291 + 0.98560028*d)
	i := (int64(maCoef0Nano) + d.julianDayNumber*maCoef1Nano) % (360 * 1e9)
	ma := rad * (float64(i)/1e9 + float64(d.time)/daySec*(float64(maCoef1Nano)/1e18))
	if ma < 0 {
		ma = ma + 2*math.Pi
	}
	return ma
}

func eclipticLongitude(m float64) float64 {
	c := rad * (1.9148*math.Sin(m) + 0.02*math.Sin(2*m) + 0.0003*math.Sin(3*m)) // equation of center
	p := rad * 102.9372                                                         // perihelion of the Earth

	return m + c + p + math.Pi
}

func sunCoords(d JulianDate) (dec, ra float64) {
	m := solarMeanAnomaly(d)
	l := eclipticLongitude(m)

	dec = declination(l, 0)
	ra = rightAscension(l, 0)
	return
}

type sunTime struct {
	angle    float64
	riseName string
	setName  string
}

//A SunCalc represents a object to calculate sun times from earth
type SunCalc struct {
	times []sunTime
}

// NewSunCalc returns a new SunCalc
func NewSunCalc() SunCalc {
	times := []sunTime{
		{angle: -0.833, riseName: "sunrise", setName: "sunset"},
		{angle: -0.3, riseName: "sunriseEnd", setName: "sunsetStart"},
		{angle: -6, riseName: "dawn", setName: "dusk"},
		{angle: -12, riseName: "nauticalDawn", setName: "nauticalDusk"},
		{angle: -18, riseName: "nightEnd", setName: "night"},
		{angle: 6, riseName: "goldenHourEnd", setName: "goldenHour"},
	}
	return SunCalc{
		times: times,
	}
}

// GetPosition calculates sun position for a given date and latitude/longitude
func (s *SunCalc) GetPosition(date time.Time, lat, lng float64) (azim, alti float64) {
	lw := rad * -lng
	phi := rad * lat
	d := toDays(date)

	dec, ra := sunCoords(d)
	h := siderealTime(d, lw) - ra

	azim = azimuth(h, phi, dec)
	alti = altitude(h, phi, dec)
	return
}

// AddTime adds times to calc
func (s *SunCalc) AddTime(angle float64, riseName, setName string) {
	s.times = append(s.times, sunTime{angle: angle, riseName: riseName, setName: setName})
}

// calculations for sun times

const j0 = 0.0009

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func julianCycle(d JulianDate, lw float64) float64 {
	return round(float64(d.julianDayNumber)+float64(d.time)/(daySec*1e9)-j0-lw/(2*math.Pi), 0.5, 0)
}

func approxTransit(Ht, lw, n float64) float64 {
	return j0 + (Ht+lw)/(2*math.Pi) + n
}

func solarTransitJ(ds, M, L float64) float64 {
	return j2000 + ds + 0.0053*math.Sin(M) - 0.0069*math.Sin(2*L)
}

func hourAngle(h, phi, d float64) float64 {
	return math.Acos((math.Sin(h) - math.Sin(phi)*math.Sin(d)) / (math.Cos(phi) * math.Cos(d)))
}

// returns set time for the given sun altitude
func getSetJ(h, lw, phi, dec, n, m, l float64) float64 {
	w := hourAngle(h, phi, dec)
	a := approxTransit(w, lw, n)
	return solarTransitJ(a, m, l)
}

// GetTimes calculates sun times for a given date and latitude/longitude
func (s *SunCalc) GetTimes(date time.Time, lat, lng float64) map[string]time.Time {

	lw := rad * -lng
	phi := rad * lat
	d := toDays(date)
	n := julianCycle(d, lw)
	ds := approxTransit(0, lw, n)

	m := solarMeanAnomaly(JulianFromFloat(ds))
	l := eclipticLongitude(m)
	dec := declination(l, 0)

	jNoon := solarTransitJ(ds, m, l)

	result := make(map[string]time.Time)
	result["solarNoon"] = JulianFromFloat(jNoon).Time()
	result["nadir"] = JulianFromFloat(jNoon - 0.5).Time()
	for _, t := range s.times {
		jSet := getSetJ(t.angle*rad, lw, phi, dec, n, m, l)
		jRise := jNoon - (jSet - jNoon)
		if t.riseName != "" {
			result[t.riseName] = JulianFromFloat(jRise).Time()
		}
		if t.setName != "" {
			result[t.setName] = JulianFromFloat(jSet).Time()
		}
	}

	return result
}

// moon calculations, based on http://aa.quae.nl/en/reken/hemelpositie.html formulas

func moonCoords(jd JulianDate) (dec, ra, dist float64) { // geocentric ecliptic coordinates of the moon
	d := float64(jd.julianDayNumber) + float64(jd.time)/(1e9*daySec)

	L := rad * (218.316 + 13.176396*d) // ecliptic longitude
	M := rad * (134.963 + 13.064993*d) // mean anomaly
	F := rad * (93.272 + 13.229350*d)  // mean distance

	l := L + rad*6.289*math.Sin(M)   // longitude
	b := rad * 5.128 * math.Sin(F)   // latitude
	dt := 385001 - 20905*math.Cos(M) // distance to the moon in km

	ra = rightAscension(l, b)
	dec = declination(l, b)
	dist = dt
	return
}

// GetMoonPosition returns the following properties:
//  alti: moon altitude above the horizon in radians
//  azim: moon azimuth in radians
//  dist: distance to moon in kilometers
func GetMoonPosition(date time.Time, lat, lng float64) (azim, alti, dist float64) {

	lw := rad * -lng
	phi := rad * lat
	d := toDays(date)

	dec, ra, distance := moonCoords(d)
	H := siderealTime(d, lw) - ra
	h := altitude(H, phi, dec)

	// altitude correction for refraction
	h = h + rad*0.017/math.Tan(h+rad*10.26/(h+rad*5.10))

	azim = azimuth(H, phi, dec)
	alti = h
	dist = distance
	return
}

const sdist = 149598000 // distance from Earth to Sun in km
// calculations for illumination parameters of the moon,
// based on http://idlastro.gsfc.nasa.gov/ftp/pro/astro/mphase.pro formulas and
// Chapter 48 of "Astronomical Algorithms" 2nd edition by Jean Meeus (Willmann-Bell, Richmond) 1998.

// GetMoonIllumination returns an the following properties:
//  fraction: illuminated fraction of the moon; varies from `0.0` (new moon) to `1.0` (full moon)
//  phase: moon phase; varies from `0.0` to `1.0`, described below
//  angle: midpoint angle in radians of the illuminated limb of the moon reckoned eastward from the north point of the disk;
// the moon is waxing if the angle is negative, and waning if positive
func GetMoonIllumination(date time.Time) (fraction, phase, angle float64) {
	d := toDays(date)
	sDec, sRa := sunCoords(d)
	mDec, mRa, mDist := moonCoords(d)

	phi := math.Acos(math.Sin(sDec)*math.Sin(mDec) + math.Cos(sDec)*math.Cos(mDec)*math.Cos(sRa-mRa))
	inc := math.Atan2(sdist*math.Sin(phi), mDist-sdist*math.Cos(phi))
	angle = math.Atan2(math.Cos(sDec)*math.Sin(sRa-mRa), math.Sin(sDec)*math.Cos(mDec)-math.Cos(sDec)*math.Sin(mDec)*math.Cos(sRa-mRa))

	fraction = (1 + math.Cos(inc)) / 2
	c := 1.0
	if angle < 0 {
		c = -1.0
	}
	phase = 0.5 + 0.5*inc*c/math.Pi

	return
}
