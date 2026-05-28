package tests

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alparslanahmed/keyword-market-mcp/internal/clustering"
)

func TestClusterKeywordsGroupsByTheme(t *testing.T) {
	items := []clustering.KeywordForCluster{
		{Keyword: "coffee maker", AvgMonthlySearches: i64(1000), CompetitionIndex: f64(20), OpportunityScore: f64(70)},
		{Keyword: "coffee grinder", AvgMonthlySearches: i64(500), CompetitionIndex: f64(40), OpportunityScore: f64(60)},
		{Keyword: "espresso machine", AvgMonthlySearches: i64(2000), CompetitionIndex: f64(60), OpportunityScore: f64(50)},
		{Keyword: "espresso maker", AvgMonthlySearches: i64(800), CompetitionIndex: f64(30), OpportunityScore: f64(65)},
	}
	clusters := clustering.ClusterKeywords(items, "English")

	names := make([]string, 0, len(clusters))
	total := 0
	for _, c := range clusters {
		names = append(names, c.Name)
		total += c.KeywordCount
	}
	sort.Strings(names)
	assert.Equal(t, []string{"coffee", "espresso"}, names)
	assert.Equal(t, 4, total)

	for _, c := range clusters {
		if c.Name == "coffee" {
			assert.Equal(t, 2, c.KeywordCount)
			assert.Equal(t, int64(1500), c.TotalSearchVolume)
			assert.NotNil(t, c.MedianCompetitionIndex)
			assert.InDelta(t, 30.0, *c.MedianCompetitionIndex, 0.001)
			assert.InDelta(t, 65.0, c.AverageOpportunityScore, 0.001)
		}
	}
}

func TestClusterKeywordsDeterministic(t *testing.T) {
	items := []clustering.KeywordForCluster{
		{Keyword: "cheap espresso maker"},
		{Keyword: "best espresso maker"},
		{Keyword: "cheap coffee grinder"},
	}
	a := clustering.ClusterKeywords(items, "English")
	b := clustering.ClusterKeywords(items, "English")
	assert.Equal(t, a, b)
}

func TestClusterKeywordsEmpty(t *testing.T) {
	assert.Empty(t, clustering.ClusterKeywords(nil, "English"))
}
