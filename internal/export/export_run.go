package export

import (
	"encoding/json"
	"time"

	"github.com/alparslanahmed/keyword-market-mcp/internal/db"
	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// Supported export formats.
const (
	FormatJSON = "json"
	FormatCSV  = "csv"
)

// RunExport is the structured, client-safe representation of a research run.
type RunExport struct {
	RunID                 string          `json:"runId"`
	Idea                  string          `json:"idea"`
	Country               string          `json:"country"`
	Language              string          `json:"language"`
	Status                string          `json:"status"`
	Verdict               string          `json:"verdict"`
	DemandScore           float64         `json:"demandScore"`
	CompetitionScore      float64         `json:"competitionScore"`
	CommercialIntentScore float64         `json:"commercialIntentScore"`
	OpportunityScore      float64         `json:"opportunityScore"`
	CreatedAt             string          `json:"createdAt"`
	CompletedAt           *string         `json:"completedAt"`
	Summary               json.RawMessage `json:"summary,omitempty"`
	Keywords              []KeywordRow    `json:"keywords"`
}

// KeywordRow is one keyword line in an export.
type KeywordRow struct {
	Keyword               string   `json:"keyword"`
	Intent                string   `json:"intent"`
	Cluster               string   `json:"cluster"`
	SourceSeed            string   `json:"sourceSeed"`
	OpportunityScore      float64  `json:"opportunityScore"`
	VolumeScore           float64  `json:"volumeScore"`
	CompetitionScore      float64  `json:"competitionScore"`
	CommercialIntentScore float64  `json:"commercialIntentScore"`
	AvgMonthlySearches    *int64   `json:"avgMonthlySearches"`
	Competition           string   `json:"competition"`
	CompetitionIndex      *float64 `json:"competitionIndex"`
}

// BuildRunExport assembles a RunExport from stored data.
func BuildRunExport(store *db.Store, runID string) (*RunExport, error) {
	run, err := store.GetRun(runID)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, utils.NewValidationInputError("Research run not found: " + runID)
	}

	rks, err := store.ListRunKeywords(runID)
	if err != nil {
		return nil, err
	}
	keywordIDs := make([]string, 0, len(rks))
	metricIDs := make([]string, 0, len(rks))
	for _, rk := range rks {
		keywordIDs = append(keywordIDs, rk.KeywordID)
		if rk.MetricID != "" {
			metricIDs = append(metricIDs, rk.MetricID)
		}
	}
	kwMap, err := store.GetKeywordsByIDs(keywordIDs)
	if err != nil {
		return nil, err
	}
	mMap, err := store.GetMetricsByIDs(metricIDs)
	if err != nil {
		return nil, err
	}

	exp := &RunExport{
		RunID:                 run.ID,
		Idea:                  run.Idea,
		Country:               run.Country,
		Language:              run.Language,
		Status:                run.Status,
		Verdict:               run.Verdict,
		DemandScore:           run.DemandScore,
		CompetitionScore:      run.CompetitionScore,
		CommercialIntentScore: run.CommercialIntentScore,
		OpportunityScore:      run.OpportunityScore,
		CreatedAt:             run.CreatedAt.UTC().Format(time.RFC3339),
		Keywords:              make([]KeywordRow, 0, len(rks)),
	}
	if run.CompletedAt != nil {
		c := run.CompletedAt.UTC().Format(time.RFC3339)
		exp.CompletedAt = &c
	}
	if run.SummaryJSON != "" {
		exp.Summary = json.RawMessage(run.SummaryJSON)
	}

	for _, rk := range rks {
		row := KeywordRow{
			Keyword:               kwMap[rk.KeywordID].Text,
			Intent:                rk.Intent,
			Cluster:               rk.ClusterName,
			SourceSeed:            rk.SourceSeed,
			OpportunityScore:      rk.OpportunityScore,
			VolumeScore:           rk.VolumeScore,
			CompetitionScore:      rk.CompetitionScore,
			CommercialIntentScore: rk.CommercialIntentScore,
		}
		if m, ok := mMap[rk.MetricID]; ok {
			row.AvgMonthlySearches = m.AvgMonthlySearches
			row.Competition = m.Competition
			row.CompetitionIndex = m.CompetitionIndex
		}
		exp.Keywords = append(exp.Keywords, row)
	}
	return exp, nil
}
