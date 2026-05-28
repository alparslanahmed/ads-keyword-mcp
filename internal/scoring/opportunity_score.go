package scoring

import (
	"fmt"
	"math"
	"strings"

	"github.com/alparslanahmed/keyword-market-mcp/internal/googleads"
	"github.com/alparslanahmed/keyword-market-mcp/internal/intent"
	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// Score holds the opportunity score and its component breakdown for one keyword.
type Score struct {
	Keyword               string  `json:"keyword"`
	OpportunityScore      float64 `json:"opportunityScore"`
	VolumeScore           float64 `json:"volumeScore"`
	CompetitionScore      float64 `json:"competitionScore"`
	CommercialIntentScore float64 `json:"commercialIntentScore"`
	TrendScore            float64 `json:"trendScore"`
	LongTailBonus         float64 `json:"longTailBonus"`
	BroadnessPenalty      float64 `json:"broadnessPenalty"`
	Intent                string  `json:"intent"`
	Explanation           string  `json:"explanation"`
}

// Reference points for log-scaling. These are directional, not calibrated to a
// specific currency or market.
const (
	volumeReference = 100_000.0 // monthly searches mapped to ~100
	bidReferenceUSD = 10.0      // top-of-page bid (currency units) mapped to ~100
)

// ScoreKeyword computes the opportunity score (0-100) and component breakdown
// for a single keyword metric.
func ScoreKeyword(m googleads.KeywordMetric, language string) Score {
	words := utils.WordCount(m.Keyword)
	kwIntent := intent.ClassifyIntent(m.Keyword, language)

	volumeScore := volumeComponent(m.AvgMonthlySearches)
	competitionScore := competitionComponent(m)
	commercialScore := commercialComponent(m, kwIntent)
	trend := TrendScore(m.MonthlySearchVolumes)
	longTail := longTailBonus(words)
	broadness := broadnessPenalty(words)

	opportunity := 0.40*volumeScore + 0.35*competitionScore + 0.25*commercialScore
	opportunity += longTail - broadness + trend
	opportunity = clamp(opportunity, 0, 100)

	s := Score{
		Keyword:               m.Keyword,
		OpportunityScore:      round1(opportunity),
		VolumeScore:           round1(volumeScore),
		CompetitionScore:      round1(competitionScore),
		CommercialIntentScore: round1(commercialScore),
		TrendScore:            round1(trend),
		LongTailBonus:         round1(longTail),
		BroadnessPenalty:      round1(broadness),
		Intent:                kwIntent,
	}
	s.Explanation = explain(s, m)
	return s
}

func volumeComponent(avg *int64) float64 {
	if avg == nil || *avg <= 0 {
		return 0
	}
	v := float64(*avg)
	score := math.Log10(v+1) / math.Log10(volumeReference+1) * 100
	return clamp(score, 0, 100)
}

func competitionComponent(m googleads.KeywordMetric) float64 {
	if m.CompetitionIndex != nil {
		return clamp(100-*m.CompetitionIndex, 0, 100)
	}
	switch m.Competition {
	case googleads.CompetitionLow:
		return 80
	case googleads.CompetitionMedium:
		return 50
	case googleads.CompetitionHigh:
		return 20
	default:
		return 50
	}
}

func commercialComponent(m googleads.KeywordMetric, kwIntent string) float64 {
	score := bidComponent(m.HighTopOfPageBidMicros)
	if intent.HasBuyingModifier(m.Keyword) {
		if score < 60 {
			score = 60
		}
		score += 15
	}
	switch kwIntent {
	case intent.Commercial:
		score += 20
	case intent.Comparison:
		score += 10
	}
	return clamp(score, 0, 100)
}

func bidComponent(highBidMicros *int64) float64 {
	if highBidMicros == nil || *highBidMicros <= 0 {
		return 0
	}
	units := utils.MicrosToUnits(highBidMicros)
	score := math.Log10(*units+1) / math.Log10(bidReferenceUSD+1) * 100
	return clamp(score, 0, 100)
}

func longTailBonus(words int) float64 {
	if words < 3 {
		return 0
	}
	return math.Min(float64(words-2)*3, 10)
}

func broadnessPenalty(words int) float64 {
	if words <= 1 {
		return 10
	}
	return 0
}

func explain(s Score, m googleads.KeywordMetric) string {
	var parts []string
	switch {
	case s.VolumeScore >= 70:
		parts = append(parts, "high search volume")
	case s.VolumeScore >= 40:
		parts = append(parts, "moderate search volume")
	case m.AvgMonthlySearches == nil:
		parts = append(parts, "no volume data")
	default:
		parts = append(parts, "low search volume")
	}
	switch {
	case s.CompetitionScore >= 70:
		parts = append(parts, "low competition")
	case s.CompetitionScore >= 45:
		parts = append(parts, "medium competition")
	default:
		parts = append(parts, "high competition")
	}
	if s.CommercialIntentScore >= 60 {
		parts = append(parts, "strong commercial intent")
	}
	if s.LongTailBonus > 0 {
		parts = append(parts, "long-tail")
	}
	if s.BroadnessPenalty > 0 {
		parts = append(parts, "broad single-word term")
	}
	if s.TrendScore >= 5 {
		parts = append(parts, "rising trend")
	} else if s.TrendScore <= -5 {
		parts = append(parts, "declining trend")
	}
	return fmt.Sprintf("%s (intent: %s)", strings.Join(parts, ", "), s.Intent)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}
