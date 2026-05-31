package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// MaxKeywordLength bounds a single normalized keyword. Anything longer is
// treated as junk and dropped.
const MaxKeywordLength = 80

// NormalizeKeyword lowercases (Turkish-aware), trims, collapses whitespace,
// strips punctuation, and applies Unicode NFC normalization. Turkish letters
// (ı İ ş ğ ü ö ç) are preserved rather than ASCII-folded.
func NormalizeKeyword(s string) string {
	if s == "" {
		return ""
	}
	// Apply canonical composition first so combining sequences fold to single runes.
	s = norm.NFC.String(s)
	s = turkishToLower(s)

	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			b.WriteRune(r)
			prevSpace = false
		case unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r):
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
		default:
			// Drop control and other characters entirely.
		}
	}
	return strings.TrimSpace(b.String())
}

// turkishToLower lowercases while keeping the dotted/dotless I distinction sane.
// The capital dotted "İ" (U+0130) is mapped to "i" directly to avoid the
// combining-dot artifact produced by strings.ToLower, and every listed Turkish
// lowercase letter is preserved.
func turkishToLower(s string) string {
	s = strings.ReplaceAll(s, "İ", "i")
	return strings.ToLower(s)
}

// IsValidKeyword reports whether a normalized keyword is usable.
func IsValidKeyword(s string) bool {
	n := NormalizeKeyword(s)
	if n == "" {
		return false
	}
	if len([]rune(n)) > MaxKeywordLength {
		return false
	}
	// Reject keywords with no letters/digits at all (already stripped, so any
	// remaining content must contain at least one).
	for _, r := range n {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return true
		}
	}
	return false
}

// NormalizeKeywords normalizes a slice, drops empties/junk, and deduplicates by
// normalized form while preserving first-seen ordering.
func NormalizeKeywords(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, raw := range in {
		n := NormalizeKeyword(raw)
		if n == "" || len([]rune(n)) > MaxKeywordLength {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

// Tokenize splits a keyword into normalized tokens.
func Tokenize(s string) []string {
	n := NormalizeKeyword(s)
	if n == "" {
		return nil
	}
	return strings.Fields(n)
}

// WordCount returns the number of whitespace-separated tokens in the normalized
// keyword.
func WordCount(s string) int {
	return len(Tokenize(s))
}
