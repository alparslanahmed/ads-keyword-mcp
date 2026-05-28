package validation

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alparslanahmed/keyword-market-mcp/internal/clustering"
	"github.com/alparslanahmed/keyword-market-mcp/internal/googleads"
	"github.com/alparslanahmed/keyword-market-mcp/internal/intent"
)

// Insights is the qualitative output derived deterministically from the data.
type Insights struct {
	Risks                 []string
	RecommendedAngles     []string
	NextResearchQuestions []string
	OneParagraphSummary   string
}

const directionalDisclaimer = "Keyword volume is a directional signal, not proof of sales or product-market fit."

// BuildInsights produces careful, factual risks/angles/questions and a summary.
// The wording deliberately avoids overclaiming (no "this proves demand").
func BuildInsights(idea string, t googleads.Targeting, s VerdictSignals, verdict, confidence string, clusters []clustering.Cluster, scored []ScoredKeyword) Insights {
	return Insights{
		Risks:                 buildRisks(s),
		RecommendedAngles:     buildAngles(clusters, scored),
		NextResearchQuestions: buildQuestions(t, s, scored),
		OneParagraphSummary:   buildSummary(idea, t, s, verdict, confidence),
	}
}

func buildRisks(s VerdictSignals) []string {
	var risks []string
	if s.KeywordsWithVolume == 0 {
		risks = append(risks, "Keyword Planner returned little or no volume data, so demand is hard to confirm.")
	}
	if s.DemandScore < 35 {
		risks = append(risks, "Search demand appears limited based on Keyword Planner data.")
	}
	if s.CompetitionScore < 40 {
		risks = append(risks, "Competition appears high across many keywords, which can raise customer acquisition costs.")
	}
	if s.CommercialKeywords == 0 {
		risks = append(risks, "Few commercial-intent keywords were found; searchers may not be in a buying mindset.")
	}
	if s.ClusterCount <= 1 {
		risks = append(risks, "Keywords cluster narrowly, suggesting a thin or undifferentiated market.")
	}
	risks = append(risks, directionalDisclaimer)
	return risks
}

func buildAngles(clusters []clustering.Cluster, scored []ScoredKeyword) []string {
	var angles []string

	// Low-competition, high-opportunity keywords make good entry points.
	lowComp := make([]ScoredKeyword, 0, len(scored))
	for _, sk := range scored {
		if sk.Score.CompetitionScore >= 60 {
			lowComp = append(lowComp, sk)
		}
	}
	sort.SliceStable(lowComp, func(i, j int) bool {
		return lowComp[i].Score.OpportunityScore > lowComp[j].Score.OpportunityScore
	})
	for i := 0; i < len(lowComp) && i < 2; i++ {
		angles = append(angles, fmt.Sprintf("Target the low-competition keyword %q as an entry point.", lowComp[i].Metric.Keyword))
	}

	for i := 0; i < len(clusters) && i < 2; i++ {
		c := clusters[i]
		angles = append(angles, fmt.Sprintf("Build content and landing pages around the %q cluster (%d keywords).", c.Name, c.KeywordCount))
	}

	// Surface a representative commercial keyword for positioning.
	for _, sk := range scored {
		if sk.Score.Intent == intent.Commercial {
			angles = append(angles, fmt.Sprintf("Emphasize commercial-intent terms like %q on product pages.", sk.Metric.Keyword))
			break
		}
	}

	if len(angles) == 0 {
		angles = append(angles, "Refine the idea into more specific, higher-intent keywords before committing.")
	}
	return angles
}

func buildQuestions(t googleads.Targeting, s VerdictSignals, scored []ScoredKeyword) []string {
	questions := []string{
		fmt.Sprintf("Do these searches reflect buyers in %s, or broader informational interest?", t.Country),
		"Can you validate willingness to pay (e.g., a landing page plus a small ads test) for the top clusters?",
		"How seasonal is demand? Review the monthly search-volume trend before committing.",
	}
	if s.CompetitionScore < 50 {
		for _, sk := range scored {
			if sk.Score.CompetitionScore < 40 {
				questions = append(questions, fmt.Sprintf("Who are the main competitors ranking or bidding for %q?", sk.Metric.Keyword))
				break
			}
		}
	}
	return questions
}

func buildSummary(idea string, t googleads.Targeting, s VerdictSignals, verdict, confidence string) string {
	demandWord := band(s.DemandScore, "limited", "moderate", "healthy")
	compWord := band(s.CompetitionScore, "high", "moderate", "low")
	return fmt.Sprintf(
		"Search demand appears %s and competition looks %s for %q in %s (%s). Keyword Planner data suggests a %q opportunity (confidence: %s) across roughly %d related keywords in %d clusters. %s",
		demandWord, compWord, strings.TrimSpace(idea), t.Country, t.Language,
		verdict, confidence, s.KeywordCount, s.ClusterCount, directionalDisclaimer,
	)
}

func band(v float64, low, mid, high string) string {
	switch {
	case v >= 60:
		return high
	case v >= 35:
		return mid
	default:
		return low
	}
}
