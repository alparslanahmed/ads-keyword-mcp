package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alparslanahmed/keyword-market-mcp/internal/googleads"
	"github.com/alparslanahmed/keyword-market-mcp/internal/intent"
	"github.com/alparslanahmed/keyword-market-mcp/internal/scoring"
)

func i64(v int64) *int64     { return &v }
func f64(v float64) *float64 { return &v }

func TestScoreKeywordHighOpportunity(t *testing.T) {
	m := googleads.KeywordMetric{
		Keyword:                "best portable espresso maker",
		AvgMonthlySearches:     i64(10000),
		Competition:            googleads.CompetitionLow,
		CompetitionIndex:       f64(10),
		HighTopOfPageBidMicros: i64(2_000_000),
	}
	s := scoring.ScoreKeyword(m, "English")
	assert.Greater(t, s.OpportunityScore, 70.0)
	assert.InDelta(t, 80.0, s.VolumeScore, 2.0)
	assert.InDelta(t, 90.0, s.CompetitionScore, 0.1)
	assert.Equal(t, intent.Comparison, s.Intent)
	assert.Greater(t, s.LongTailBonus, 0.0)
	assert.Equal(t, 0.0, s.BroadnessPenalty)
}

func TestScoreKeywordNilVolume(t *testing.T) {
	s := scoring.ScoreKeyword(googleads.KeywordMetric{Keyword: "espresso maker", Competition: googleads.CompetitionUnknown}, "English")
	assert.Equal(t, 0.0, s.VolumeScore)
}

func TestScoreKeywordHighCompetitionPenalized(t *testing.T) {
	low := scoring.ScoreKeyword(googleads.KeywordMetric{Keyword: "espresso maker", AvgMonthlySearches: i64(5000), CompetitionIndex: f64(10), Competition: googleads.CompetitionLow}, "English")
	high := scoring.ScoreKeyword(googleads.KeywordMetric{Keyword: "espresso maker", AvgMonthlySearches: i64(5000), CompetitionIndex: f64(90), Competition: googleads.CompetitionHigh}, "English")
	assert.Greater(t, low.CompetitionScore, high.CompetitionScore)
	assert.Greater(t, low.OpportunityScore, high.OpportunityScore)
}

func TestScoreKeywordBroadnessPenalty(t *testing.T) {
	s := scoring.ScoreKeyword(googleads.KeywordMetric{Keyword: "espresso", AvgMonthlySearches: i64(5000), CompetitionIndex: f64(50)}, "English")
	assert.Equal(t, 10.0, s.BroadnessPenalty)
}

func TestTrendScore(t *testing.T) {
	rising := []googleads.MonthlySearchVolume{
		{Year: 2024, Month: 1, MonthlySearches: 100}, {Year: 2024, Month: 2, MonthlySearches: 100},
		{Year: 2024, Month: 3, MonthlySearches: 100}, {Year: 2024, Month: 4, MonthlySearches: 200},
		{Year: 2024, Month: 5, MonthlySearches: 200}, {Year: 2024, Month: 6, MonthlySearches: 200},
	}
	assert.Equal(t, 10.0, scoring.TrendScore(rising))

	declining := []googleads.MonthlySearchVolume{
		{Year: 2024, Month: 1, MonthlySearches: 200}, {Year: 2024, Month: 2, MonthlySearches: 200},
		{Year: 2024, Month: 3, MonthlySearches: 200}, {Year: 2024, Month: 4, MonthlySearches: 100},
		{Year: 2024, Month: 5, MonthlySearches: 100}, {Year: 2024, Month: 6, MonthlySearches: 100},
	}
	assert.Equal(t, -10.0, scoring.TrendScore(declining))

	// Insufficient data returns no signal.
	assert.Equal(t, 0.0, scoring.TrendScore(rising[:3]))
}
