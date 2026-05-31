package validation

import (
	"strings"

	"github.com/alparslanahmed/keyword-market-mcp/internal/clustering"
	"github.com/alparslanahmed/keyword-market-mcp/internal/googleads"
	"github.com/alparslanahmed/keyword-market-mcp/internal/scoring"
	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// ScoreKeywordInput is one keyword for score_keyword_opportunities. Bid values
// are supplied in currency units (not micros) and converted internally.
type ScoreKeywordInput struct {
	Keyword              string       `json:"keyword"`
	AvgMonthlySearches   *int64       `json:"avgMonthlySearches"`
	Competition          *string      `json:"competition"`
	CompetitionIndex     *float64     `json:"competitionIndex"`
	LowTopOfPageBid      *float64     `json:"lowTopOfPageBid"`
	HighTopOfPageBid     *float64     `json:"highTopOfPageBid"`
	MonthlySearchVolumes []MonthlyOut `json:"monthlySearchVolumes"`
}

// ScoreKeywords scores caller-supplied metrics without any Google Ads call.
func ScoreKeywords(inputs []ScoreKeywordInput, language string) []ScoreOut {
	out := make([]ScoreOut, 0, len(inputs))
	for _, in := range inputs {
		m := googleads.KeywordMetric{
			Keyword:                strings.TrimSpace(in.Keyword),
			AvgMonthlySearches:     in.AvgMonthlySearches,
			Competition:            normalizeCompetitionPtr(in.Competition),
			CompetitionIndex:       in.CompetitionIndex,
			LowTopOfPageBidMicros:  unitsPtrToMicros(in.LowTopOfPageBid),
			HighTopOfPageBidMicros: unitsPtrToMicros(in.HighTopOfPageBid),
			MonthlySearchVolumes:   fromMonthlyOut(in.MonthlySearchVolumes),
		}
		sc := scoring.ScoreKeyword(m, language)
		out = append(out, ScoreOut{
			Keyword:               sc.Keyword,
			OpportunityScore:      sc.OpportunityScore,
			VolumeScore:           sc.VolumeScore,
			CompetitionScore:      sc.CompetitionScore,
			CommercialIntentScore: sc.CommercialIntentScore,
			TrendScore:            sc.TrendScore,
			LongTailBonus:         sc.LongTailBonus,
			Intent:                sc.Intent,
			Explanation:           sc.Explanation,
		})
	}
	return out
}

// ClusterKeywordInput is one keyword for the cluster_keywords tool.
type ClusterKeywordInput struct {
	Keyword            string   `json:"keyword"`
	AvgMonthlySearches *int64   `json:"avgMonthlySearches"`
	CompetitionIndex   *float64 `json:"competitionIndex"`
	OpportunityScore   *float64 `json:"opportunityScore"`
	Intent             *string  `json:"intent"`
}

// ClusterKeywordsTool clusters caller-supplied keywords without any API call.
func ClusterKeywordsTool(inputs []ClusterKeywordInput, language string) []clustering.Cluster {
	items := make([]clustering.KeywordForCluster, 0, len(inputs))
	for _, in := range inputs {
		intentVal := ""
		if in.Intent != nil {
			intentVal = *in.Intent
		}
		items = append(items, clustering.KeywordForCluster{
			Keyword:            in.Keyword,
			AvgMonthlySearches: in.AvgMonthlySearches,
			CompetitionIndex:   in.CompetitionIndex,
			OpportunityScore:   in.OpportunityScore,
			Intent:             intentVal,
		})
	}
	return clustering.ClusterKeywords(items, language)
}

func normalizeCompetitionPtr(s *string) string {
	if s == nil {
		return googleads.CompetitionUnknown
	}
	return googleads.NormalizeCompetition(strings.ToUpper(strings.TrimSpace(*s)))
}

func unitsPtrToMicros(units *float64) *int64 {
	if units == nil {
		return nil
	}
	v := utils.UnitsToMicros(*units)
	return &v
}

func fromMonthlyOut(in []MonthlyOut) []googleads.MonthlySearchVolume {
	out := make([]googleads.MonthlySearchVolume, 0, len(in))
	for _, m := range in {
		out = append(out, googleads.MonthlySearchVolume{Year: m.Year, Month: m.Month, MonthlySearches: m.MonthlySearches})
	}
	return out
}
