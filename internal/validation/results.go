package validation

import (
	"time"

	"github.com/alparslanahmed/keyword-market-mcp/internal/clustering"
	"github.com/alparslanahmed/keyword-market-mcp/internal/googleads"
	"github.com/alparslanahmed/keyword-market-mcp/internal/scoring"
	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// MonthlyOut is an output monthly volume point.
type MonthlyOut struct {
	Year            int   `json:"year"`
	Month           int   `json:"month"`
	MonthlySearches int64 `json:"monthlySearches"`
}

// KeywordIdeaOut is one normalized keyword idea (discover_keywords output).
type KeywordIdeaOut struct {
	Keyword              string       `json:"keyword"`
	AvgMonthlySearches   *int64       `json:"avgMonthlySearches"`
	Competition          string       `json:"competition"`
	CompetitionIndex     *float64     `json:"competitionIndex"`
	LowTopOfPageBid      *float64     `json:"lowTopOfPageBid"`
	HighTopOfPageBid     *float64     `json:"highTopOfPageBid"`
	MonthlySearchVolumes []MonthlyOut `json:"monthlySearchVolumes"`
	SourceSeed           string       `json:"sourceSeed"`
	CacheHit             bool         `json:"cacheHit"`
}

// MetricOut is one normalized historical metric (get_keyword_metrics output).
type MetricOut struct {
	Keyword              string       `json:"keyword"`
	AvgMonthlySearches   *int64       `json:"avgMonthlySearches"`
	MonthlySearchVolumes []MonthlyOut `json:"monthlySearchVolumes"`
	Competition          string       `json:"competition"`
	CompetitionIndex     *float64     `json:"competitionIndex"`
	LowTopOfPageBid      *float64     `json:"lowTopOfPageBid"`
	HighTopOfPageBid     *float64     `json:"highTopOfPageBid"`
	FetchedAt            string       `json:"fetchedAt"`
	CacheHit             bool         `json:"cacheHit"`
}

// DiscoverResult is the discover_keywords tool output.
type DiscoverResult struct {
	Keywords []KeywordIdeaOut `json:"keywords"`
}

// MetricsResult is the get_keyword_metrics tool output.
type MetricsResult struct {
	Metrics []MetricOut `json:"metrics"`
}

// TargetingOut mirrors resolved targeting in outputs.
type TargetingOut struct {
	Country           string `json:"country"`
	Language          string `json:"language"`
	GeoTargetConstant string `json:"geoTargetConstant"`
	LanguageConstant  string `json:"languageConstant"`
}

// SummaryOut is the validation summary block.
type SummaryOut struct {
	Verdict               string  `json:"verdict"`
	Confidence            string  `json:"confidence"`
	DemandScore           float64 `json:"demandScore"`
	CompetitionScore      float64 `json:"competitionScore"`
	CommercialIntentScore float64 `json:"commercialIntentScore"`
	OpportunityScore      float64 `json:"opportunityScore"`
	OneParagraphSummary   string  `json:"oneParagraphSummary"`
}

// OpportunityOut is one row in topOpportunities.
type OpportunityOut struct {
	Keyword            string   `json:"keyword"`
	AvgMonthlySearches *int64   `json:"avgMonthlySearches"`
	Competition        string   `json:"competition"`
	CompetitionIndex   *float64 `json:"competitionIndex"`
	LowTopOfPageBid    *float64 `json:"lowTopOfPageBid"`
	HighTopOfPageBid   *float64 `json:"highTopOfPageBid"`
	OpportunityScore   float64  `json:"opportunityScore"`
	Intent             string   `json:"intent"`
	Cluster            string   `json:"cluster"`
}

// ValidateResult is the validate_product_idea tool output.
type ValidateResult struct {
	Idea                  string               `json:"idea"`
	Targeting             TargetingOut         `json:"targeting"`
	Summary               SummaryOut           `json:"summary"`
	TopOpportunities      []OpportunityOut     `json:"topOpportunities"`
	Clusters              []clustering.Cluster `json:"clusters"`
	Risks                 []string             `json:"risks"`
	RecommendedAngles     []string             `json:"recommendedAngles"`
	NextResearchQuestions []string             `json:"nextResearchQuestions"`
	RunID                 string               `json:"runId"`
}

// RankedIdeaOut is one entry in a comparison ranking.
type RankedIdeaOut struct {
	Idea                  string   `json:"idea"`
	Verdict               string   `json:"verdict"`
	OpportunityScore      float64  `json:"opportunityScore"`
	DemandScore           float64  `json:"demandScore"`
	CompetitionScore      float64  `json:"competitionScore"`
	CommercialIntentScore float64  `json:"commercialIntentScore"`
	BestClusters          []string `json:"bestClusters"`
	MainRisks             []string `json:"mainRisks"`
	RunID                 string   `json:"runId"`
}

// CompareResult is the compare_product_ideas tool output.
type CompareResult struct {
	RankedIdeas []RankedIdeaOut `json:"rankedIdeas"`
	Winner      *string         `json:"winner"`
	Summary     string          `json:"summary"`
}

// ScoreOut is one row of score_keyword_opportunities output.
type ScoreOut struct {
	Keyword               string  `json:"keyword"`
	OpportunityScore      float64 `json:"opportunityScore"`
	VolumeScore           float64 `json:"volumeScore"`
	CompetitionScore      float64 `json:"competitionScore"`
	CommercialIntentScore float64 `json:"commercialIntentScore"`
	TrendScore            float64 `json:"trendScore"`
	LongTailBonus         float64 `json:"longTailBonus"`
	Intent                string  `json:"intent"`
	Explanation           string  `json:"explanation"`
}

func toMonthlyOut(in []googleads.MonthlySearchVolume) []MonthlyOut {
	out := make([]MonthlyOut, 0, len(in))
	for _, m := range in {
		out = append(out, MonthlyOut{Year: m.Year, Month: m.Month, MonthlySearches: m.MonthlySearches})
	}
	return out
}

func toIdeaOut(idea googleads.KeywordIdea, cacheHit bool) KeywordIdeaOut {
	return KeywordIdeaOut{
		Keyword:              idea.Keyword,
		AvgMonthlySearches:   idea.AvgMonthlySearches,
		Competition:          googleads.NormalizeCompetition(idea.Competition),
		CompetitionIndex:     idea.CompetitionIndex,
		LowTopOfPageBid:      utils.MicrosToUnits(idea.LowTopOfPageBidMicros),
		HighTopOfPageBid:     utils.MicrosToUnits(idea.HighTopOfPageBidMicros),
		MonthlySearchVolumes: toMonthlyOut(idea.MonthlySearchVolumes),
		SourceSeed:           idea.SourceSeed,
		CacheHit:             cacheHit,
	}
}

func toMetricOut(m googleads.KeywordMetric, fetchedAt time.Time, cacheHit bool) MetricOut {
	return MetricOut{
		Keyword:              m.Keyword,
		AvgMonthlySearches:   m.AvgMonthlySearches,
		MonthlySearchVolumes: toMonthlyOut(m.MonthlySearchVolumes),
		Competition:          googleads.NormalizeCompetition(m.Competition),
		CompetitionIndex:     m.CompetitionIndex,
		LowTopOfPageBid:      utils.MicrosToUnits(m.LowTopOfPageBidMicros),
		HighTopOfPageBid:     utils.MicrosToUnits(m.HighTopOfPageBidMicros),
		FetchedAt:            fetchedAt.UTC().Format(time.RFC3339),
		CacheHit:             cacheHit,
	}
}

// ScoredKeyword pairs a keyword metric with its computed score and cluster.
type ScoredKeyword struct {
	Metric     googleads.KeywordMetric
	Score      scoring.Score
	Cluster    string
	SourceSeed string
}

func toOpportunityOut(sk ScoredKeyword) OpportunityOut {
	return OpportunityOut{
		Keyword:            sk.Metric.Keyword,
		AvgMonthlySearches: sk.Metric.AvgMonthlySearches,
		Competition:        googleads.NormalizeCompetition(sk.Metric.Competition),
		CompetitionIndex:   sk.Metric.CompetitionIndex,
		LowTopOfPageBid:    utils.MicrosToUnits(sk.Metric.LowTopOfPageBidMicros),
		HighTopOfPageBid:   utils.MicrosToUnits(sk.Metric.HighTopOfPageBidMicros),
		OpportunityScore:   sk.Score.OpportunityScore,
		Intent:             sk.Score.Intent,
		Cluster:            sk.Cluster,
	}
}
