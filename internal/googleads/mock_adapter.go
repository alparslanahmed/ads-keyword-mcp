package googleads

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"strings"
	"time"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// MockGoogleAdsAdapter produces deterministic, realistic-looking keyword data
// without any credentials or network access. Metrics are derived from a hash of
// the keyword so results are stable across runs (important for tests) while
// still varying per keyword. It is enabled by APP_ENV=test or GOOGLE_ADS_MOCK=true.
type MockGoogleAdsAdapter struct {
	now func() time.Time
}

// NewMockGoogleAdsAdapter constructs a mock adapter.
func NewMockGoogleAdsAdapter() *MockGoogleAdsAdapter {
	return &MockGoogleAdsAdapter{now: time.Now}
}

// expansion templates used to derive related keyword ideas from each seed.
var (
	mockTemplatesEN = []string{
		"%s", "best %s", "%s price", "buy %s", "%s review", "%s reviews",
		"cheap %s", "%s online", "portable %s", "%s near me", "%s alternative",
		"how to use %s",
	}
	mockTemplatesTR = []string{
		"%s fiyat", "en iyi %s", "%s yorumları", "%s satın al", "ucuz %s",
		"%s online", "%s tavsiye",
	}
)

// GenerateKeywordIdeas expands seeds (and any URL host tokens) into a
// deterministic set of keyword ideas.
func (m *MockGoogleAdsAdapter) GenerateKeywordIdeas(_ context.Context, input GenerateKeywordIdeasInput) (*GenerateKeywordIdeasOutput, error) {
	seeds := append([]string{}, input.SeedKeywords...)
	if input.URL != "" {
		if host := utils.MustHost(input.URL); host != "" {
			seeds = append(seeds, urlHostSeed(host))
		}
	}
	if len(seeds) == 0 {
		return nil, utils.NewValidationInputError("No seed keywords or URL provided.")
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 300
	}

	seen := make(map[string]struct{})
	out := &GenerateKeywordIdeasOutput{}
	templates := append(append([]string{}, mockTemplatesEN...), mockTemplatesTR...)

	for _, seed := range seeds {
		seed = strings.TrimSpace(seed)
		if seed == "" {
			continue
		}
		for _, tmpl := range templates {
			text := strings.TrimSpace(fmt.Sprintf(tmpl, seed))
			norm := utils.NormalizeKeyword(text)
			if norm == "" {
				continue
			}
			if _, ok := seen[norm]; ok {
				continue
			}
			seen[norm] = struct{}{}
			metric := m.genMetric(text)
			out.Ideas = append(out.Ideas, KeywordIdea{KeywordMetric: metric, SourceSeed: seed})
			if len(out.Ideas) >= limit {
				return out, nil
			}
		}
	}
	return out, nil
}

// GenerateHistoricalMetrics returns deterministic metrics for each keyword.
func (m *MockGoogleAdsAdapter) GenerateHistoricalMetrics(_ context.Context, input GenerateHistoricalMetricsInput) (*GenerateHistoricalMetricsOutput, error) {
	out := &GenerateHistoricalMetricsOutput{}
	for _, kw := range input.Keywords {
		if strings.TrimSpace(kw) == "" {
			continue
		}
		out.Metrics = append(out.Metrics, m.genMetric(kw))
	}
	if len(out.Metrics) == 0 {
		return nil, utils.NewNoKeywordDataError("No keywords provided for historical metrics.")
	}
	return out, nil
}

func urlHostSeed(host string) string {
	host = strings.TrimPrefix(host, "www.")
	if i := strings.IndexByte(host, '.'); i > 0 {
		host = host[:i]
	}
	return host
}

// genMetric deterministically derives a KeywordMetric from the keyword text.
func (m *MockGoogleAdsAdapter) genMetric(keyword string) KeywordMetric {
	norm := utils.NormalizeKeyword(keyword)
	h := hashString(norm)

	metric := KeywordMetric{
		Keyword: norm,
		Source:  SourceMock,
	}

	// ~4% of keywords have no volume data, to exercise nil handling.
	if h%25 == 7 {
		metric.Competition = CompetitionUnknown
		return metric
	}

	avg := int64(50 + h%9950) // 50..9999
	metric.AvgMonthlySearches = &avg

	idx := float64(h % 101) // 0..100
	metric.CompetitionIndex = &idx
	switch {
	case idx < 34:
		metric.Competition = CompetitionLow
	case idx < 67:
		metric.Competition = CompetitionMedium
	default:
		metric.Competition = CompetitionHigh
	}

	low := int64((h%500)*1000 + 50000)
	high := low + int64((h%800)*1000+100000)
	metric.LowTopOfPageBidMicros = &low
	metric.HighTopOfPageBidMicros = &high

	metric.MonthlySearchVolumes = m.genMonthly(avg, h)
	return metric
}

// genMonthly builds 24 months of volumes ending at the current month, with a
// deterministic trend (down/flat/up) and a mild seasonal wave.
func (m *MockGoogleAdsAdapter) genMonthly(avg int64, h uint64) []MonthlySearchVolume {
	const months = 24
	trend := int(h%3) - 1 // -1, 0, or 1
	base := float64(avg)
	now := m.now()
	out := make([]MonthlySearchVolume, 0, months)
	start := now.AddDate(0, -(months - 1), 0)
	for i := 0; i < months; i++ {
		t := start.AddDate(0, i, 0)
		seasonal := 1 + 0.15*math.Sin(float64(int(t.Month()))/12*2*math.Pi)
		trendFactor := 1 + float64(trend)*0.02*float64(i)
		v := int64(base * seasonal * trendFactor)
		if v < 0 {
			v = 0
		}
		out = append(out, MonthlySearchVolume{
			Year:            t.Year(),
			Month:           int(t.Month()),
			MonthlySearches: v,
		})
	}
	return out
}

func hashString(s string) uint64 {
	hh := fnv.New64a()
	_, _ = hh.Write([]byte(s))
	return hh.Sum64()
}
