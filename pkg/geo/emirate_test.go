package geo_test

import (
	"testing"

	"github.com/uae-search-oss/uae-search-oss/pkg/geo"
)

func TestClassifyCoord(t *testing.T) {
	tests := []struct {
		name     string
		lat, lng float64
		want     geo.Emirate
	}{
		{"dubai bur", 25.27, 55.29, geo.Dubai},          // Bur Dubai - center of Dubai
		{"abudhabi island", 24.46, 54.32, geo.AbuDhabi}, // Yas Island area
		{"sharjah heart", 25.42, 55.50, geo.Sharjah},    // Deep in Sharjah
		{"outside uae", 0.0, 0.0, geo.Unknown},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := geo.ClassifyCoord(tc.lat, tc.lng)
			if got != tc.want {
				t.Errorf("ClassifyCoord(%f, %f) = %q, want %q", tc.lat, tc.lng, got, tc.want)
			}
		})
	}
}

func TestDistanceKm(t *testing.T) {
	tests := []struct {
		name       string
		lat1, lng1 float64
		lat2, lng2 float64
		min, max   float64
	}{
		{"dubai to abu dhabi", 25.2, 55.3, 24.5, 54.4, 100, 200},
		{"same point", 25.2, 55.3, 25.2, 55.3, 0, 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := geo.DistanceKm(tc.lat1, tc.lng1, tc.lat2, tc.lng2)
			if got < tc.min || got > tc.max {
				t.Errorf("DistanceKm() = %f, want %f-%f", got, tc.min, tc.max)
			}
		})
	}
}

func TestGeoBoost(t *testing.T) {
	tests := []struct {
		name     string
		distance float64
		wantMin  float64
		wantMax  float64
	}{
		{"within 5km", 3, 0.95, 1.05},
		{"at 10km", 10, 0.8, 1.0},
		{"at 50km", 50, -0.05, 0.05},
		{"beyond 50km", 100, -0.1, 0.1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := geo.GeoBoost(tc.distance)
			if got < tc.wantMin || got > tc.wantMax {
				t.Errorf("GeoBoost(%f) = %f, expected %f-%f", tc.distance, got, tc.wantMin, tc.wantMax)
			}
		})
	}
}
