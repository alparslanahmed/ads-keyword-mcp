package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alparslanahmed/keyword-market-mcp/internal/cache"
)

func TestKeywordIdeasKeyOrderInsensitive(t *testing.T) {
	a := cache.KeywordIdeasKey([]string{"b keyword", "a keyword"}, "", "Turkey", "Turkish", "GOOGLE_SEARCH")
	b := cache.KeywordIdeasKey([]string{"a keyword", "b keyword"}, "", "Turkey", "Turkish", "GOOGLE_SEARCH")
	assert.Equal(t, a, b)
}

func TestKeywordIdeasKeyVariesByInputs(t *testing.T) {
	base := cache.KeywordIdeasKey([]string{"espresso"}, "", "Turkey", "Turkish", "GOOGLE_SEARCH")
	assert.NotEqual(t, base, cache.KeywordIdeasKey([]string{"espresso"}, "", "United States", "Turkish", "GOOGLE_SEARCH"))
	assert.NotEqual(t, base, cache.KeywordIdeasKey([]string{"espresso"}, "", "Turkey", "English", "GOOGLE_SEARCH"))
	assert.NotEqual(t, base, cache.KeywordIdeasKey([]string{"espresso"}, "https://x.com", "Turkey", "Turkish", "GOOGLE_SEARCH"))
	assert.NotEqual(t, base, cache.KeywordIdeasKey([]string{"latte"}, "", "Turkey", "Turkish", "GOOGLE_SEARCH"))
}

func TestHistoricalMetricsKey(t *testing.T) {
	a := cache.HistoricalMetricsKey([]string{"espresso", "latte"}, "Turkey", "Turkish", "GOOGLE_SEARCH")
	b := cache.HistoricalMetricsKey([]string{"latte", "espresso"}, "Turkey", "Turkish", "GOOGLE_SEARCH")
	assert.Equal(t, a, b)
	assert.NotEqual(t, a, cache.HistoricalMetricsKey([]string{"espresso"}, "Turkey", "Turkish", "GOOGLE_SEARCH"))
	// Idea and historical keys for the same inputs must not collide.
	assert.NotEqual(t, a, cache.KeywordIdeasKey([]string{"espresso", "latte"}, "", "Turkey", "Turkish", "GOOGLE_SEARCH"))
}
