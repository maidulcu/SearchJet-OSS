// Package geo provides UAE emirate boundaries and geo-aware ranking helpers.
package geo

import "math"

// Emirate represents one of the seven UAE emirates.
type Emirate string

const (
	AbuDhabi  Emirate = "abudhabi"
	Dubai     Emirate = "dubai"
	Sharjah   Emirate = "sharjah"
	Ajman     Emirate = "ajman"
	UmmAlQuwain Emirate = "ummalquwain"
	RasAlKhaimah Emirate = "rasalkhaimah"
	Fujairah  Emirate = "fujairah"
	Unknown   Emirate = ""
)

// BoundingBox is a simple lat/lng bounding box.
type BoundingBox struct {
	MinLat, MaxLat float64
	MinLng, MaxLng float64
}

// emirateBounds maps each emirate to its approximate bounding box.
// These are simplified rectangles suitable for initial filtering.
var emirateBounds = map[Emirate]BoundingBox{
	AbuDhabi:     {22.6, 25.2, 51.5, 56.4},
	Dubai:        {24.7, 25.4, 54.9, 55.6},
	Sharjah:      {25.1, 25.7, 55.3, 56.3},
	Ajman:        {25.3, 25.5, 55.4, 55.6},
	UmmAlQuwain:  {25.5, 25.7, 55.5, 55.8},
	RasAlKhaimah: {25.5, 26.2, 55.7, 56.3},
	Fujairah:     {25.0, 25.5, 56.2, 56.4},
}

// ClassifyCoord returns the emirate for the given lat/lng, or Unknown.
func ClassifyCoord(lat, lng float64) Emirate {
	for emirate, box := range emirateBounds {
		if lat >= box.MinLat && lat <= box.MaxLat &&
			lng >= box.MinLng && lng <= box.MaxLng {
			return emirate
		}
	}
	return Unknown
}

// DistanceKm calculates the great-circle distance in km between two coordinates
// using the Haversine formula.
func DistanceKm(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKm = 6371.0
	dLat := toRad(lat2 - lat1)
	dLng := toRad(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

// GeoBoost returns a ranking boost factor (0.0–1.0) based on distance.
// Documents within 5km get full boost; beyond 50km get no boost.
func GeoBoost(distanceKm float64) float64 {
	if distanceKm <= 5 {
		return 1.0
	}
	if distanceKm >= 50 {
		return 0.0
	}
	return 1.0 - (distanceKm-5)/45.0
}

func toRad(deg float64) float64 {
	return deg * math.Pi / 180
}
