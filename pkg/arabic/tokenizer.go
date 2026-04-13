// Package arabic provides UAE dialect tokenizer, bilingual normalization, and transliteration.
package arabic

import (
	"strings"
	"unicode"
)

var (
	commonToEmirati = map[string]string{
		"\u0644\u0645\u0627\u0630\u0627": "\u0644\u064a\u0634",             // لماذا -> ليش
		"\u0647\u0630\u0627":             "\u0647\u0627\u0644",             // هذا -> هال
		"\u0647\u0630\u0647":             "\u0647\u0630\u064a",             // هذه ->ذي
		"\u062c\u064a\u062f":             "\u0632\u064a\u0646",             // جيد -> زين
		"\u0641\u0642\u0637":             "\u0628\u0633",                   // فقط -> بس
		"\u062f\u0639\u0646\u064a":       "\u062e\u0644",                   // دعني -> خل
		"\u0645\u0631\u062d\u0628\u0627": "\u0645\u0631\u0631",             // مرحبا -> مرر
		"\u0645\u0627":                   "\u0648\u0634",                   // ما ->وش
		"\u0627\u0646\u062a\u0645":       "\u0627\u0646\u062a\u0648",       // أنتم ->أنتو
		"\u0637\u0628\u064a\u0639\u064a": "\u0639\u0627\u062f\u064a",       // طبيعي -> عادي
		"\u064a\u062d\u062f\u062b":       "\u064a\u062a\u062d\u0642\u0642", // يحدث -> يتحقق
	}

	transliterationMap = map[rune]rune{
		'\u0627': 'a', '\u0628': 'b', '\u062a': 't', '\u062b': 't',
		'\u062c': 'j', '\u062d': 'h', '\u062e': 'k', '\u062f': 'd',
		'\u0630': 'd', '\u0631': 'r', '\u0632': 'z', '\u0633': 's',
		'\u0634': 's', '\u0635': 's', '\u0636': 'd', '\u0637': 't',
		'\u0638': 'z', '\u0639': 'a', '\u063a': 'g', '\u0641': 'f',
		'\u0642': 'q', '\u0643': 'k', '\u0644': 'l', '\u0645': 'm',
		'\u0646': 'n', '\u0647': 'h', '\u0648': 'w', '\u064a': 'y',
	}
)

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

func Tokenize(text string) []string {
	normalized := Normalize(text)
	parts := strings.FieldsFunc(normalized, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	tokens := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) > 1 {
			tokens = append(tokens, strings.ToLower(p))
		}
	}
	return tokens
}

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

var stopWordsAR = map[string]bool{
	"\u0641\u064a": true, "\u0645\u0646": true, "\u0639\u0644\u0649": true, "\u0625\u0644\u0649": true, "\u0639\u0646": true,
	"\u0645\u0639": true, "\u0647\u0630\u0627": true, "\u0647\u0630\u0647": true, "\u0627\u0644\u062a\u064a": true, "\u0627\u0644\u0630\u064a": true,
	"\u0643\u0627\u0646": true, "\u0643\u0627\u0646\u062a": true, "\u064a\u0643\u0648\u0646": true, "\u0648\u0647\u0648": true, "\u0648\u0647\u064a": true,
}

var stopWordsEN = map[string]bool{
	"the": true, "a": true, "an": true, "in": true, "on": true,
	"at": true, "to": true, "for": true, "of": true, "and": true,
	"or": true, "but": true, "is": true, "are": true, "was": true,
	"were": true, "be": true, "been": true, "being": true,
}

func RemoveStopWords(tokens []string) []string {
	result := tokens[:0]
	for _, t := range tokens {
		if !stopWordsAR[t] && !stopWordsEN[t] {
			result = append(result, t)
		}
	}
	return result
}

func normalizeRune(r rune) rune {
	if r >= 0x064B && r <= 0x065F || r == 0x0670 {
		return 0
	}
	if r == 0x0640 {
		return 0
	}
	switch r {
	case 'أ', 'إ', 'آ', 'ٱ':
		return 'ا'
	case 'ى':
		return 'ي'
	case 'ة':
		return 'ه'
	}
	return r
}

func NormalizeBilingual(text string) string {
	lang := DetectLang(text)

	if lang == "ar" || lang == "mixed" {
		return Normalize(text)
	}

	return strings.ToLower(strings.TrimSpace(text))
}

func ExpandQuery(query string) []string {
	lang := DetectLang(query)
	tokens := Tokenize(query)

	if lang == "ar" || lang == "mixed" {
		expanded := make([]string, 0, len(tokens)*2)
		for _, t := range tokens {
			expanded = append(expanded, t)
			if substitute, ok := commonToEmirati[t]; ok {
				expanded = append(expanded, substitute)
			}
		}
		return expanded
	}

	return tokens
}

func Transliterate(arabic string) string {
	var result strings.Builder
	for _, r := range arabic {
		if replacement, ok := transliterationMap[r]; ok {
			result.WriteRune(replacement)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
