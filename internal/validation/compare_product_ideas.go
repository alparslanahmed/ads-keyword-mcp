package validation

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// CompareInput is the compare_product_ideas request.
type CompareInput struct {
	Ideas                  []string
	Country                string
	Language               string
	MaxKeywordIdeasPerIdea int
	ForceRefresh           bool
}

// CompareProductIdeas validates each idea (reusing the full flow, so caching and
// rate limiting apply) and returns a ranked comparison.
func (s *Service) CompareProductIdeas(ctx context.Context, in CompareInput) (*CompareResult, error) {
	ideas := dedupeNonEmpty(in.Ideas)
	if len(ideas) == 0 {
		return nil, utils.NewValidationInputError("Provide at least one product idea to compare.")
	}
	perIdea := in.MaxKeywordIdeasPerIdea
	if perIdea <= 0 {
		perIdea = 150
	}

	ranked := make([]RankedIdeaOut, 0, len(ideas))
	for _, idea := range ideas {
		res, err := s.ValidateProductIdea(ctx, ValidateInput{
			Idea:            idea,
			Country:         in.Country,
			Language:        in.Language,
			MaxKeywordIdeas: perIdea,
			ForceRefresh:    in.ForceRefresh,
		})
		if err != nil {
			ranked = append(ranked, RankedIdeaOut{
				Idea:      idea,
				Verdict:   VerdictInconclusive,
				MainRisks: []string{utils.AsAppError(err).Message},
			})
			continue
		}
		ranked = append(ranked, RankedIdeaOut{
			Idea:                  res.Idea,
			Verdict:               res.Summary.Verdict,
			OpportunityScore:      res.Summary.OpportunityScore,
			DemandScore:           res.Summary.DemandScore,
			CompetitionScore:      res.Summary.CompetitionScore,
			CommercialIntentScore: res.Summary.CommercialIntentScore,
			BestClusters:          clusterNames(res, 3),
			MainRisks:             firstN(res.Risks, 2),
			RunID:                 res.RunID,
		})
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].OpportunityScore != ranked[j].OpportunityScore {
			return ranked[i].OpportunityScore > ranked[j].OpportunityScore
		}
		return ranked[i].DemandScore > ranked[j].DemandScore
	})

	var winner *string
	if len(ranked) > 0 && ranked[0].Verdict != VerdictInconclusive {
		w := ranked[0].Idea
		winner = &w
	}

	return &CompareResult{
		RankedIdeas: ranked,
		Winner:      winner,
		Summary:     comparisonSummary(ranked, winner),
	}, nil
}

func comparisonSummary(ranked []RankedIdeaOut, winner *string) string {
	if len(ranked) == 0 {
		return "No ideas were compared."
	}
	if winner == nil {
		return "None of the compared ideas showed a clear opportunity in Keyword Planner data; results are inconclusive."
	}
	top := ranked[0]
	return fmt.Sprintf(
		"%q ranks highest with an opportunity score of %.1f (%s) versus %d other idea(s). This reflects keyword demand signals, not confirmed sales.",
		*winner, top.OpportunityScore, top.Verdict, len(ranked)-1,
	)
}

func clusterNames(res *ValidateResult, n int) []string {
	out := make([]string, 0, n)
	for i, c := range res.Clusters {
		if i >= n {
			break
		}
		out = append(out, c.Name)
	}
	return out
}

func firstN(in []string, n int) []string {
	if len(in) <= n {
		return in
	}
	return in[:n]
}

func dedupeNonEmpty(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		trimmed := strings.TrimSpace(s)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
