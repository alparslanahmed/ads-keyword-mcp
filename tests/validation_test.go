package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alparslanahmed/keyword-market-mcp/internal/export"
	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
	"github.com/alparslanahmed/keyword-market-mcp/internal/validation"
)

func strptr(s string) *string { return &s }

func TestDetermineVerdict(t *testing.T) {
	strong := validation.VerdictSignals{KeywordCount: 60, KeywordsWithVolume: 30, CommercialKeywords: 5, LowMedCompetitionKeywords: 10, ClusterCount: 4, DemandScore: 70}
	assert.Equal(t, validation.VerdictStrong, validation.DetermineVerdict(strong))

	promising := validation.VerdictSignals{KeywordCount: 20, KeywordsWithVolume: 10, CommercialKeywords: 2, LowMedCompetitionKeywords: 3, ClusterCount: 2, DemandScore: 45}
	assert.Equal(t, validation.VerdictPromising, validation.DetermineVerdict(promising))

	weak := validation.VerdictSignals{KeywordCount: 20, KeywordsWithVolume: 10, CommercialKeywords: 0, LowMedCompetitionKeywords: 0, ClusterCount: 1, DemandScore: 20}
	assert.Equal(t, validation.VerdictWeak, validation.DetermineVerdict(weak))

	assert.Equal(t, validation.VerdictInconclusive, validation.DetermineVerdict(validation.VerdictSignals{KeywordCount: 5, KeywordsWithVolume: 2}))
	assert.Equal(t, validation.VerdictInconclusive, validation.DetermineVerdict(validation.VerdictSignals{KeywordCount: 40, KeywordsWithVolume: 0}))
}

func TestDetermineConfidence(t *testing.T) {
	assert.Equal(t, validation.ConfidenceHigh, validation.DetermineConfidence(validation.VerdictSignals{KeywordCount: 60, KeywordsWithVolume: 25}))
	assert.Equal(t, validation.ConfidenceMedium, validation.DetermineConfidence(validation.VerdictSignals{KeywordCount: 20, KeywordsWithVolume: 8}))
	assert.Equal(t, validation.ConfidenceLow, validation.DetermineConfidence(validation.VerdictSignals{KeywordCount: 5, KeywordsWithVolume: 1}))
}

func TestValidateProductIdeaEndToEnd(t *testing.T) {
	st := newTestStack(t)
	res, err := st.service.ValidateProductIdea(context.Background(), validation.ValidateInput{
		Idea:            "portable espresso maker for campers",
		Country:         "Turkey",
		Language:        "Turkish",
		MaxKeywordIdeas: 120,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, res.RunID)
	assert.Contains(t, []string{
		validation.VerdictStrong, validation.VerdictPromising, validation.VerdictWeak, validation.VerdictInconclusive,
	}, res.Summary.Verdict)
	assert.NotEmpty(t, res.TopOpportunities)
	assert.NotEmpty(t, res.Clusters)
	assert.NotEmpty(t, res.Risks)
	assert.Equal(t, "geoTargetConstants/2792", res.Targeting.GeoTargetConstant)

	// The run is persisted and exportable.
	exp, err := export.BuildRunExport(st.store, res.RunID)
	require.NoError(t, err)
	assert.NotEmpty(t, exp.Keywords)

	csvText, err := export.ToCSV(exp)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(csvText, "keyword,intent,cluster"))

	runs, err := st.store.ListRecentRuns(10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(runs), 1)
}

func TestValidateProductIdeaInvalidInputs(t *testing.T) {
	st := newTestStack(t)
	_, err := st.service.ValidateProductIdea(context.Background(), validation.ValidateInput{Idea: ""})
	assert.Equal(t, utils.CodeValidationInput, utils.AsAppError(err).Code)

	_, err = st.service.ValidateProductIdea(context.Background(), validation.ValidateInput{Idea: "x", Country: "Atlantis"})
	assert.Equal(t, utils.CodeInvalidTargeting, utils.AsAppError(err).Code)
}

func TestCompareProductIdeas(t *testing.T) {
	st := newTestStack(t)
	res, err := st.service.CompareProductIdeas(context.Background(), validation.CompareInput{
		Ideas:                  []string{"portable espresso maker", "camping coffee grinder"},
		Country:                "Turkey",
		Language:               "Turkish",
		MaxKeywordIdeasPerIdea: 80,
	})
	require.NoError(t, err)
	assert.Len(t, res.RankedIdeas, 2)
	// Ranking is sorted by opportunity score descending.
	assert.GreaterOrEqual(t, res.RankedIdeas[0].OpportunityScore, res.RankedIdeas[1].OpportunityScore)
}

func TestScoreKeywordsTool(t *testing.T) {
	out := validation.ScoreKeywords([]validation.ScoreKeywordInput{
		{Keyword: "best espresso maker", AvgMonthlySearches: i64(5000), Competition: strptr("LOW"), CompetitionIndex: f64(20), HighTopOfPageBid: f64(2)},
	}, "English")
	require.Len(t, out, 1)
	assert.Greater(t, out[0].OpportunityScore, 0.0)
	assert.Equal(t, "comparison", out[0].Intent)
}

func TestClusterKeywordsTool(t *testing.T) {
	out := validation.ClusterKeywordsTool([]validation.ClusterKeywordInput{
		{Keyword: "cheap espresso maker", AvgMonthlySearches: i64(100)},
		{Keyword: "cheap coffee grinder", AvgMonthlySearches: i64(200)},
		{Keyword: "best espresso maker", AvgMonthlySearches: i64(300)},
	}, "English")
	assert.NotEmpty(t, out)
}

func TestDiscoverAndMetrics(t *testing.T) {
	st := newTestStack(t)
	disc, err := st.service.DiscoverKeywords(context.Background(), validation.DiscoverInput{
		SeedKeywords: []string{"espresso maker"}, Country: "Turkey", Language: "Turkish", Limit: 30,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, disc.Keywords)

	// Second call should hit the cache.
	disc2, err := st.service.DiscoverKeywords(context.Background(), validation.DiscoverInput{
		SeedKeywords: []string{"espresso maker"}, Country: "Turkey", Language: "Turkish", Limit: 30,
	})
	require.NoError(t, err)
	require.NotEmpty(t, disc2.Keywords)
	assert.True(t, disc2.Keywords[0].CacheHit)

	met, err := st.service.GetKeywordMetrics(context.Background(), validation.MetricsInput{
		Keywords: []string{"espresso maker", "coffee grinder"}, Country: "Turkey", Language: "Turkish",
	})
	require.NoError(t, err)
	assert.Len(t, met.Metrics, 2)
}
