// Package prayertime provides UAE prayer times for contextual ranking.
// Uses static calculation based on UAE location; can be extended with API.
package prayertime

import (
	"math"
	"time"
)

type Prayer int

const (
	Fajr Prayer = iota
	Sunrise
	Zuhr
	Asr
	Maghrib
	Isha
)

var prayerNames = []string{"Fajr", "Sunrise", "Zuhr", "Asr", "Maghrib", "Isha"}

func (p Prayer) String() string {
	return prayerNames[p]
}

type PrayerTimes struct {
	Date         time.Time
	Emirate      string
	Fajr         time.Time
	Sunrise      time.Time
	Zuhr         time.Time
	Asr          time.Time
	Maghrib      time.Time
	Isha         time.Time
	FajrAdhan    time.Time
	MaghribAdhan time.Time
}

type Location struct {
	Name     string
	Lat      float64
	Lng      float64
	Timezone string
}

var uaeLocations = map[string]Location{
	"dubai":        {"Dubai", 25.2048, 55.2708, "Asia/Dubai"},
	"abudhabi":     {"Abu Dhabi", 24.4539, 54.3773, "Asia/Dubai"},
	"sharjah":      {"Sharjah", 25.3463, 55.4209, "Asia/Dubai"},
	"ajman":        {"Ajman", 25.4052, 55.5136, "Asia/Dubai"},
	"rasalkhaimah": {"Ras Al Khaimah", 25.7895, 55.9432, "Asia/Dubai"},
	"fujairah":     {"Fujairah", 25.1285, 56.3265, "Asia/Dubai"},
	"ummalquwain":  {"Umm Al Quwain", 25.5647, 55.5553, "Asia/Dubai"},
}

func GetLocation(emirate string) (Location, bool) {
	l, ok := uaeLocations[emirate]
	return l, ok
}

func CalculatePrayerTimes(date time.Time, emirate string) (*PrayerTimes, error) {
	loc, ok := GetLocation(emirate)
	if !ok {
		loc = uaeLocations["dubai"]
	}

	lat := loc.Lat
	dayOfYear := date.YearDay()

	declination := 23.45 * math.Sin(2*math.Pi*(float64(dayOfYear)-81)/365)
	hourAngle := math.Acos(-math.Tan(lat*math.Pi/180)*math.Tan(declination*math.Pi/180)) * 180 / math.Pi

	sunriseUTC := 12 - hourAngle/15
	fajrUTC := 12 - (hourAngle+15)/15
	zuhrUTC := 12.0
	asrUTC := 12 + hourAngle/10
	maghribUTC := 12 + hourAngle/15
	ishaUTC := 12 + (hourAngle+18)/15

	pt := &PrayerTimes{
		Date:    date,
		Emirate: emirate,
	}

	locTz := time.UTC
	if loc.Timezone != "" {
		if tz, err := time.LoadLocation(loc.Timezone); err == nil {
			locTz = tz
		}
	}

	pt.Fajr = time.Date(date.Year(), date.Month(), date.Day(), int(fajrUTC), int((fajrUTC-float64(int(fajrUTC)))*60), 0, 0, locTz)
	pt.Sunrise = time.Date(date.Year(), date.Month(), date.Day(), int(sunriseUTC), int((sunriseUTC-float64(int(sunriseUTC)))*60), 0, 0, locTz)
	pt.Zuhr = time.Date(date.Year(), date.Month(), date.Day(), int(zuhrUTC), 0, 0, 0, locTz)
	pt.Asr = time.Date(date.Year(), date.Month(), date.Day(), int(asrUTC), int((asrUTC - float64(int(asrUTC))*60)), 0, 0, locTz)
	pt.Maghrib = time.Date(date.Year(), date.Month(), date.Day(), int(maghribUTC), int((maghribUTC - float64(int(maghribUTC))*60)), 0, 0, locTz)
	pt.Isha = time.Date(date.Year(), date.Month(), date.Day(), int(ishaUTC), int((ishaUTC - float64(int(ishaUTC))*60)), 0, 0, locTz)

	return pt, nil
}

func (pt *PrayerTimes) GetCurrentPrayer() Prayer {
	now := time.Now()

	if now.Before(pt.Fajr) || now.After(pt.Isha) {
		return Isha
	}
	if now.Before(pt.Sunrise) {
		return Fajr
	}
	if now.Before(pt.Zuhr) {
		return Sunrise
	}
	if now.Before(pt.Asr) {
		return Zuhr
	}
	if now.Before(pt.Maghrib) {
		return Asr
	}
	if now.Before(pt.Isha) {
		return Maghrib
	}

	return Isha
}

func (pt *PrayerTimes) TimeUntilNextPrayer() time.Duration {
	now := time.Now()

	switch pt.GetCurrentPrayer() {
	case Fajr:
		return pt.Fajr.Sub(now)
	case Sunrise:
		return pt.Zuhr.Sub(now)
	case Zuhr:
		return pt.Asr.Sub(now)
	case Asr:
		return pt.Maghrib.Sub(now)
	case Maghrib:
		return pt.Isha.Sub(now)
	case Isha:
		return pt.Fajr.Add(24 * time.Hour).Sub(now)
	}
	return 0
}

func (pt *PrayerTimes) IsPrayerTime(minsBefore, minsAfter int) bool {
	now := time.Now()

	prayers := []time.Time{pt.Fajr, pt.Zuhr, pt.Asr, pt.Maghrib, pt.Isha}
	window := time.Duration(minsBefore+minsAfter) * time.Minute

	for _, p := range prayers {
		if now.Add(-time.Duration(minsBefore)*time.Minute).Before(p) &&
			now.Add(time.Duration(minsAfter)*time.Minute).After(p) {
			return true
		}
	}
	_ = window
	return false
}

func (pt *PrayerTimes) GetBoostFactor() float64 {
	if pt.IsPrayerTime(30, 30) {
		return 0.9
	}
	return 1.0
}
