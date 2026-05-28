package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

func TestNormalizeKeyword(t *testing.T) {
	cases := map[string]string{
		"  Coffee   Maker  ":   "coffee maker",
		"Coffee, Maker!!":      "coffee maker",
		"İstanbul Kahve":       "istanbul kahve",
		"Çocuk  Oyuncağı":      "çocuk oyuncağı",
		"ŞĞÜÖÇ":                "şğüöç",
		"taşınabilir espresso": "taşınabilir espresso",
		"":                     "",
		"!!!":                  "",
	}
	for in, want := range cases {
		assert.Equalf(t, want, utils.NormalizeKeyword(in), "input %q", in)
	}
}

func TestNormalizeKeywordPreservesTurkishChars(t *testing.T) {
	// Turkish-specific letters must survive normalization (not ASCII-folded).
	for _, r := range []string{"ı", "ş", "ğ", "ü", "ö", "ç"} {
		assert.Contains(t, utils.NormalizeKeyword("kahve "+r), r)
	}
	// Dotted capital İ lowercases to a plain "i" (no combining-dot artifact).
	assert.Equal(t, "ic", utils.NormalizeKeyword("İC"))
}

func TestNormalizeKeywordsDeduplicates(t *testing.T) {
	got := utils.NormalizeKeywords([]string{"Coffee", "coffee", " COFFEE ", "Tea", "tea!!"})
	assert.Equal(t, []string{"coffee", "tea"}, got)
}

func TestNormalizeKeywordsDropsJunkAndEmpties(t *testing.T) {
	got := utils.NormalizeKeywords([]string{"", "   ", "!!!", "valid keyword"})
	assert.Equal(t, []string{"valid keyword"}, got)
}

func TestIsValidKeyword(t *testing.T) {
	assert.True(t, utils.IsValidKeyword("coffee maker"))
	assert.False(t, utils.IsValidKeyword(""))
	assert.False(t, utils.IsValidKeyword("   "))
	// Over the length limit.
	long := make([]byte, 0, 200)
	for i := 0; i < 200; i++ {
		long = append(long, 'a')
	}
	assert.False(t, utils.IsValidKeyword(string(long)))
}

func TestWordCount(t *testing.T) {
	assert.Equal(t, 0, utils.WordCount(""))
	assert.Equal(t, 1, utils.WordCount("espresso"))
	assert.Equal(t, 3, utils.WordCount("portable espresso maker"))
}
