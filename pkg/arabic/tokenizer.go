// Package arabic provides a lightweight UAE Arabic tokenizer for search indexing.
// This is a pure-Go implementation suitable for basic query normalization.
// For advanced NLP (AraBERT, dialect detection), use the gRPC NLP bridge.
package arabic

import (
	"strings"
	"unicode"
)

// Normalize performs basic Arabic text normalization for UAE search:
// - Removes diacritics (tashkeel/harakat)
// - Normalizes Alef variants (أ إ آ → ا)
// - Normalizes Yaa variants (ى → ي)
// - Normalizes Taa Marbuta (ة → ه)
// - Removes tatweel (ـ)
func Normalize(text string) string {
	var b strings.Builder
	b.Grow(len(text))

	for _, r := range text {
		r = normalizeRune(r)
		if r != 0 {
			b.WriteRune(r)
		}
	}

	return b.String()
}

// Tokenize splits Arabic/English mixed text into search tokens.
// Handles right-to-left Arabic words and left-to-right Latin words.
func Tokenize(text string) []string {
	normalized := Normalize(text)
	parts := strings.FieldsFunc(normalized, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	tokens := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) > 1 { // skip single-char noise
			tokens = append(tokens, strings.ToLower(p))
		}
	}
	return tokens
}

// DetectLang returns "ar", "en", or "mixed" based on character frequency.
func DetectLang(text string) string {
	var arabicCount, latinCount int
	for _, r := range text {
		if unicode.Is(unicode.Arabic, r) {
			arabicCount++
		} else if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			latinCount++
		}
	}
	if arabicCount == 0 && latinCount == 0 {
		return "unknown"
	}
	ratio := float64(arabicCount) / float64(arabicCount+latinCount)
	switch {
	case ratio > 0.7:
		return "ar"
	case ratio < 0.3:
		return "en"
	default:
		return "mixed"
	}
}

// UAE common stop words (Arabic + transliterated names)
var stopWordsAR = map[string]bool{
	"في": true, "من": true, "على": true, "إلى": true, "عن": true,
	"مع": true, "هذا": true, "هذه": true, "التي": true, "الذي": true,
	"كان": true, "كانت": true, "يكون": true, "وهو": true, "وهي": true,
}

// RemoveStopWords filters Arabic stop words from a token list.
func RemoveStopWords(tokens []string) []string {
	result := tokens[:0]
	for _, t := range tokens {
		if !stopWordsAR[t] {
			result = append(result, t)
		}
	}
	return result
}

func normalizeRune(r rune) rune {
	// Remove Arabic diacritics (U+064B–U+065F, U+0670)
	if r >= 0x064B && r <= 0x065F || r == 0x0670 {
		return 0
	}
	// Remove tatweel
	if r == 0x0640 {
		return 0
	}
	// Normalize Alef variants → bare Alef
	switch r {
	case 'أ', 'إ', 'آ', 'ٱ':
		return 'ا'
	// Normalize Yaa variants
	case 'ى':
		return 'ي'
	// Normalize Taa Marbuta
	case 'ة':
		return 'ه'
	}
	return r
}
