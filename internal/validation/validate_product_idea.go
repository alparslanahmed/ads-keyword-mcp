package validation

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/alparslanahmed/keyword-market-mcp/internal/clustering"
	"github.com/alparslanahmed/keyword-market-mcp/internal/db"
	"github.com/alparslanahmed/keyword-market-mcp/internal/googleads"
	"github.com/alparslanahmed/keyword-market-mcp/internal/intent"
	"github.com/alparslanahmed/keyword-market-mcp/internal/scoring"
	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// aggVolumeReference scales total demand to the 0-100 demand score.
const aggVolumeReference = 500_000.0

// ValidateInput is the validate_product_idea request.
type ValidateInput struct {
	Idea                 string
	Country              string
	Language             string
	SeedKeywords         []string
	URL                  string
	Audience             string
	Problem              string
	MaxKeywordIdeas      int
	MinSearchVolume      *int64
	IncludeAdultKeywords bool
	ForceRefresh         bool
}

type analysis struct {
	scored     []ScoredKeyword
	clusters   []clustering.Cluster
	signals    VerdictSignals
	summary    SummaryOut
	verdict    string
	confidence string
	insights   Insights
}

// ValidateProductIdea runs the end-to-end validation flow and persists a
// ResearchRun.
func (s *Service) ValidateProductIdea(ctx context.Context, in ValidateInput) (*ValidateResult, error) {
	idea := strings.TrimSpace(in.Idea)
	if idea == "" {
		return nil, utils.NewValidationInputError("idea is required.")
	}
	t, err := s.resolveTargeting(in.Country, in.Language)
	if err != nil {
		return nil, err
	}
	maxIdeas := in.MaxKeywordIdeas
	if maxIdeas <= 0 {
		maxIdeas = s.defaultMaxIdeas
	}

	url := s.sanitizeURL(in.URL)
	seeds := utils.NormalizeKeywords(append(GenerateSeeds(idea, t.Language), in.SeedKeywords...))
	if len(seeds) == 0 && url == "" {
		return nil, utils.NewValidationInputError("Could not derive seed keywords from the idea; add seedKeywords or a URL.")
	}

	now := time.Now()
	seedInput, _ := json.Marshal(map[string]any{
		"seeds": seeds, "url": url, "audience": in.Audience, "problem": in.Problem,
	})
	run := &db.ResearchRun{
		Idea:          idea,
		Country:       t.Country,
		Language:      t.Language,
		SeedInputJSON: string(seedInput),
		Status:        "running",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.store.CreateRun(run); err != nil {
		return nil, err
	}

	ideas, _, err := s.fetchIdeas(ctx, t, seeds, url, maxIdeas, in.IncludeAdultKeywords, in.ForceRefresh)
	if err != nil {
		s.failRun(run, err)
		return nil, err
	}

	a := s.analyzeIdeas(idea, t, ideas, in.MinSearchVolume)
	s.persistRunResults(run, t, a, now)

	return s.buildValidateResult(idea, t, a, run.ID), nil
}

// analyzeIdeas dedupes, filters, scores, clusters, and summarizes keyword ideas.
func (s *Service) analyzeIdeas(idea string, t googleads.Targeting, ideas []googleads.KeywordIdea, minVol *int64) analysis {
	seen := make(map[string]struct{}, len(ideas))
	scored := make([]ScoredKeyword, 0, len(ideas))
	for _, kw := range ideas {
		norm := utils.NormalizeKeyword(kw.Keyword)
		if norm == "" {
			continue
		}
		if _, ok := seen[norm]; ok {
			continue
		}
		if minVol != nil {
			if kw.AvgMonthlySearches == nil || *kw.AvgMonthlySearches < *minVol {
				continue
			}
		}
		seen[norm] = struct{}{}
		sc := scoring.ScoreKeyword(kw.KeywordMetric, t.Language)
		m := kw.KeywordMetric
		m.Keyword = norm
		scored = append(scored, ScoredKeyword{Metric: m, Score: sc, SourceSeed: kw.SourceSeed})
	}

	items := make([]clustering.KeywordForCluster, len(scored))
	for i, sk := range scored {
		opp := sk.Score.OpportunityScore
		items[i] = clustering.KeywordForCluster{
			Keyword:            sk.Metric.Keyword,
			AvgMonthlySearches: sk.Metric.AvgMonthlySearches,
			CompetitionIndex:   sk.Metric.CompetitionIndex,
			OpportunityScore:   &opp,
			Intent:             sk.Score.Intent,
		}
	}
	clusters, assignment := clustering.ClusterWithAssignments(items, t.Language)
	for i := range scored {
		if i < len(assignment) {
			scored[i].Cluster = assignment[i]
		}
	}

	signals := computeSignals(scored, clusters)
	verdict := DetermineVerdict(signals)
	confidence := DetermineConfidence(signals)
	insights := BuildInsights(idea, t, signals, verdict, confidence, clusters, topScored(scored, 20))

	return analysis{
		scored:     scored,
		clusters:   clusters,
		signals:    signals,
		verdict:    verdict,
		confidence: confidence,
		insights:   insights,
		summary: SummaryOut{
			Verdict:               verdict,
			Confidence:            confidence,
			DemandScore:           signals.DemandScore,
			CompetitionScore:      signals.CompetitionScore,
			CommercialIntentScore: signals.CommercialIntentScore,
			OpportunityScore:      signals.OpportunityScore,
			OneParagraphSummary:   insights.OneParagraphSummary,
		},
	}
}

func computeSignals(scored []ScoredKeyword, clusters []clustering.Cluster) VerdictSignals {
	sig := VerdictSignals{KeywordCount: len(scored), ClusterCount: len(clusters)}
	if len(scored) == 0 {
		return sig
	}
	var totalVol int64
	var compSum, commSum float64
	for _, sk := range scored {
		if sk.Metric.AvgMonthlySearches != nil {
			sig.KeywordsWithVolume++
			totalVol += *sk.Metric.AvgMonthlySearches
		}
		if sk.Score.Intent == intent.Commercial || sk.Score.Intent == intent.Comparison {
			sig.CommercialKeywords++
		}
		if sk.Metric.Competition == googleads.CompetitionLow || sk.Metric.Competition == googleads.CompetitionMedium {
			sig.LowMedCompetitionKeywords++
		}
		compSum += sk.Score.CompetitionScore
		commSum += sk.Score.CommercialIntentScore
	}
	n := float64(len(scored))
	sig.DemandScore = round1(demandFromVolume(totalVol))
	sig.CompetitionScore = round1(compSum / n)
	sig.CommercialIntentScore = round1(commSum / n)
	sig.OpportunityScore = round1(topOpportunityAverage(scored, 10))
	return sig
}

func demandFromVolume(total int64) float64 {
	if total <= 0 {
		return 0
	}
	score := math.Log10(float64(total)+1) / math.Log10(aggVolumeReference+1) * 100
	return clampFloat(score, 0, 100)
}

func topOpportunityAverage(scored []ScoredKeyword, n int) float64 {
	if len(scored) == 0 {
		return 0
	}
	sorted := topScored(scored, n)
	var sum float64
	for _, sk := range sorted {
		sum += sk.Score.OpportunityScore
	}
	return sum / float64(len(sorted))
}

func topScored(scored []ScoredKeyword, n int) []ScoredKeyword {
	sorted := make([]ScoredKeyword, len(scored))
	copy(sorted, scored)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Score.OpportunityScore != sorted[j].Score.OpportunityScore {
			return sorted[i].Score.OpportunityScore > sorted[j].Score.OpportunityScore
		}
		return sorted[i].Metric.Keyword < sorted[j].Metric.Keyword
	})
	if n > 0 && len(sorted) > n {
		sorted = sorted[:n]
	}
	return sorted
}

func (s *Service) buildValidateResult(idea string, t googleads.Targeting, a analysis, runID string) *ValidateResult {
	top := topScored(a.scored, 15)
	opportunities := make([]OpportunityOut, 0, len(top))
	for _, sk := range top {
		opportunities = append(opportunities, toOpportunityOut(sk))
	}
	return &ValidateResult{
		Idea: idea,
		Targeting: TargetingOut{
			Country:           t.Country,
			Language:          t.Language,
			GeoTargetConstant: t.GeoTargetConstant,
			LanguageConstant:  t.LanguageConstant,
		},
		Summary:               a.summary,
		TopOpportunities:      opportunities,
		Clusters:              a.clusters,
		Risks:                 a.insights.Risks,
		RecommendedAngles:     a.insights.RecommendedAngles,
		NextResearchQuestions: a.insights.NextResearchQuestions,
		RunID:                 runID,
	}
}

func (s *Service) persistRunResults(run *db.ResearchRun, t googleads.Targeting, a analysis, now time.Time) {
	for _, sk := range a.scored {
		keywordID, metricID := s.persistMetricSnapshot(t, sk.Metric, now)
		if keywordID == "" {
			continue
		}
		notes, _ := json.Marshal(map[string]any{
			"explanation":   sk.Score.Explanation,
			"trendScore":    sk.Score.TrendScore,
			"longTailBonus": sk.Score.LongTailBonus,
		})
		_ = s.store.CreateRunKeyword(&db.ResearchRunKeyword{
			ResearchRunID:         run.ID,
			KeywordID:             keywordID,
			MetricID:              metricID,
			SourceSeed:            sk.SourceSeed,
			OpportunityScore:      sk.Score.OpportunityScore,
			VolumeScore:           sk.Score.VolumeScore,
			CompetitionScore:      sk.Score.CompetitionScore,
			CommercialIntentScore: sk.Score.CommercialIntentScore,
			Intent:                sk.Score.Intent,
			ClusterName:           sk.Cluster,
			NotesJSON:             string(notes),
			CreatedAt:             now,
		})
	}

	summaryJSON, _ := json.Marshal(a.summary)
	completed := time.Now()
	run.Status = "completed"
	run.Verdict = a.verdict
	run.DemandScore = a.signals.DemandScore
	run.CompetitionScore = a.signals.CompetitionScore
	run.CommercialIntentScore = a.signals.CommercialIntentScore
	run.OpportunityScore = a.signals.OpportunityScore
	run.SummaryJSON = string(summaryJSON)
	run.CompletedAt = &completed
	run.UpdatedAt = completed
	if err := s.store.SaveRun(run); err != nil {
		s.logger.Warn("failed to save research run")
	}
}

func (s *Service) failRun(run *db.ResearchRun, err error) {
	run.Status = "failed"
	run.ErrorMessage = utils.AsAppError(err).Message
	run.UpdatedAt = time.Now()
	if saveErr := s.store.SaveRun(run); saveErr != nil {
		s.logger.Warn("failed to save failed research run")
	}
}

func clampFloat(v, lo, hi float64) float64 {
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
