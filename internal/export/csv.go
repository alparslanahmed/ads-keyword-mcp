package export

import (
	"bytes"
	"encoding/csv"
	"strconv"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

var csvHeader = []string{
	"keyword", "intent", "cluster", "source_seed",
	"opportunity_score", "volume_score", "competition_score", "commercial_intent_score",
	"avg_monthly_searches", "competition", "competition_index",
}

// ToCSV renders a RunExport's keyword rows as CSV text.
func ToCSV(exp *RunExport) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(csvHeader); err != nil {
		return "", utils.NewInternalError(err)
	}
	for _, row := range exp.Keywords {
		record := []string{
			row.Keyword,
			row.Intent,
			row.Cluster,
			row.SourceSeed,
			formatFloat(row.OpportunityScore),
			formatFloat(row.VolumeScore),
			formatFloat(row.CompetitionScore),
			formatFloat(row.CommercialIntentScore),
			formatInt64Ptr(row.AvgMonthlySearches),
			row.Competition,
			formatFloatPtr(row.CompetitionIndex),
		}
		if err := w.Write(record); err != nil {
			return "", utils.NewInternalError(err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", utils.NewInternalError(err)
	}
	return buf.String(), nil
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', 1, 64)
}

func formatInt64Ptr(v *int64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatInt(*v, 10)
}

func formatFloatPtr(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', 1, 64)
}
