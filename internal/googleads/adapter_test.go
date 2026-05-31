package googleads

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockGenerateKeywordIdeas(t *testing.T) {
	m := NewMockGoogleAdsAdapter()
	out, err := m.GenerateKeywordIdeas(context.Background(), GenerateKeywordIdeasInput{
		SeedKeywords: []string{"espresso maker"},
		Limit:        25,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, out.Ideas)
	assert.LessOrEqual(t, len(out.Ideas), 25)
	for _, idea := range out.Ideas {
		assert.NotEmpty(t, idea.Keyword)
		assert.Equal(t, SourceMock, idea.Source)
		assert.Equal(t, "espresso maker", idea.SourceSeed)
	}
}

func TestMockRequiresSeedOrURL(t *testing.T) {
	m := NewMockGoogleAdsAdapter()
	_, err := m.GenerateKeywordIdeas(context.Background(), GenerateKeywordIdeasInput{})
	assert.Error(t, err)
}

func TestMockDeterministic(t *testing.T) {
	m := NewMockGoogleAdsAdapter()
	a := m.genMetric("portable espresso maker")
	b := m.genMetric("portable espresso maker")
	assert.Equal(t, a.AvgMonthlySearches, b.AvgMonthlySearches)
	assert.Equal(t, a.Competition, b.Competition)
}

func TestMockHistoricalMetrics(t *testing.T) {
	m := NewMockGoogleAdsAdapter()
	out, err := m.GenerateHistoricalMetrics(context.Background(), GenerateHistoricalMetricsInput{
		Keywords: []string{"espresso maker", "coffee grinder"},
	})
	require.NoError(t, err)
	assert.Len(t, out.Metrics, 2)
	assert.NotEmpty(t, out.Metrics[0].MonthlySearchVolumes)
}

func TestParseIdeasSample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "testdata", "keyword_ideas_sample.json"))
	require.NoError(t, err)
	var resp ideasResponse
	require.NoError(t, json.Unmarshal(raw, &resp))

	ideas := resp.normalize(SourceGoogleAds)
	require.NotEmpty(t, ideas)

	first := ideas[0]
	assert.Equal(t, "portable espresso maker", first.Keyword)
	require.NotNil(t, first.AvgMonthlySearches)
	assert.Equal(t, int64(8100), *first.AvgMonthlySearches)
	assert.Equal(t, CompetitionMedium, first.Competition)
	require.NotNil(t, first.CompetitionIndex)
	assert.InDelta(t, 47.0, *first.CompetitionIndex, 0.001)
	// JANUARY maps to month 1.
	require.NotEmpty(t, first.MonthlySearchVolumes)
	assert.Equal(t, 1, first.MonthlySearchVolumes[0].Month)
	assert.Equal(t, 2024, first.MonthlySearchVolumes[0].Year)
}

func TestParseHistoricalSample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "testdata", "historical_metrics_sample.json"))
	require.NoError(t, err)
	var resp historicalResponse
	require.NoError(t, json.Unmarshal(raw, &resp))

	metrics := resp.normalize(SourceGoogleAds)
	require.Len(t, metrics, 2)
	assert.Equal(t, "portable espresso maker", metrics[0].Keyword)
	assert.Equal(t, 12, len(metrics[0].MonthlySearchVolumes))
}
