package arabic_test

import (
	"testing"

	"github.com/uae-search-oss/uae-search-oss/pkg/arabic"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"alef hamza above", "أحمد", "احمد"},
		{"alef hamza below", "إمارات", "امارات"},
		{"alef madda", "آل نهيان", "ال نهيان"},
		{"taa marbuta", "مدينة", "مدينه"},
		{"yaa final", "مبنى", "مبني"},
		{"diacritics stripped", "مَدِينَة", "مدينه"},
		{"tatweel removed", "مـدينة", "مدينه"},
		{"english passthrough", "Dubai", "Dubai"},
		{"mixed", "Dubai دبي", "Dubai دبي"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := arabic.Normalize(tc.input)
			if got != tc.want {
				t.Errorf("Normalize(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestDetectLang(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"pure arabic", "مطعم دبي", "ar"},
		{"pure english", "restaurant dubai", "en"},
		{"mixed", "restaurant مطعم", "en"}, // Latin chars dominate
		{"empty", "", "unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := arabic.DetectLang(tc.input)
			if got != tc.want {
				t.Errorf("DetectLang(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
	}{
		{"arabic phrase", "مطعم في دبي", 3}, // no stop word filtering in Tokenize()
		{"english phrase", "restaurant in dubai", 3},
		{"mixed", "restaurant مطعم", 2},
		{"single chars stripped", "a b c", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := arabic.Tokenize(tc.input)
			if len(got) != tc.wantCount {
				t.Errorf("Tokenize(%q) returned %d tokens, want %d: %v", tc.input, len(got), tc.wantCount, got)
			}
		})
	}
}
