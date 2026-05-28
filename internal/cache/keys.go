package cache

import (
	"sort"
	"strings"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// Cache type tags.
const (
	TypeKeywordIdeas      = "keyword_ideas"
	TypeHistoricalMetrics = "historical_metrics"
)

// KeywordIdeasKey builds a deterministic cache key for an idea-generation
// request. The key incorporates method, normalized+sorted seeds, URL, country,
// language, and network so distinct requests never collide.
func KeywordIdeasKey(seeds []string, url, country, language, network string) string {
	return utils.SHA256Hex(append([]string{
		"GenerateKeywordIdeas",
		canonicalKeywords(seeds),
		strings.ToLower(strings.TrimSpace(url)),
	}, canonicalTargeting(country, language, network)...)...)
}

// HistoricalMetricsKey builds a deterministic cache key for a historical-metrics
// request.
func HistoricalMetricsKey(keywords []string, country, language, network string) string {
	return utils.SHA256Hex(append([]string{
		"GenerateHistoricalMetrics",
		canonicalKeywords(keywords),
		"", // no URL component for historical metrics
	}, canonicalTargeting(country, language, network)...)...)
}

func canonicalKeywords(in []string) string {
	normalized := utils.NormalizeKeywords(in)
	sort.Strings(normalized)
	return strings.Join(normalized, "|")
}

func canonicalTargeting(country, language, network string) []string {
	return []string{
		strings.ToLower(strings.TrimSpace(country)),
		strings.ToLower(strings.TrimSpace(language)),
		strings.ToUpper(strings.TrimSpace(network)),
	}
}
