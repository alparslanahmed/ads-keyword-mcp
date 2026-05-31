package scoring

import (
	"sort"

	"github.com/alparslanahmed/keyword-market-mcp/internal/googleads"
)

// TrendScore returns a small additive signal in the range [-10, 10] based on
// recent growth or decline in monthly search volumes. It needs at least six
// data points to compare the latest quarter against the prior one; otherwise it
// returns 0 (no signal).
func TrendScore(volumes []googleads.MonthlySearchVolume) float64 {
	if len(volumes) < 6 {
		return 0
	}
	sorted := make([]googleads.MonthlySearchVolume, len(volumes))
	copy(sorted, volumes)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Year != sorted[j].Year {
			return sorted[i].Year < sorted[j].Year
		}
		return sorted[i].Month < sorted[j].Month
	})

	n := len(sorted)
	recent := averageVolume(sorted[n-3:])
	prev := averageVolume(sorted[n-6 : n-3])

	if prev == 0 {
		if recent > 0 {
			return 10
		}
		return 0
	}
	pctChange := (recent - prev) / prev
	return clamp(pctChange*20, -10, 10)
}

func averageVolume(v []googleads.MonthlySearchVolume) float64 {
	if len(v) == 0 {
		return 0
	}
	var sum int64
	for _, m := range v {
		sum += m.MonthlySearches
	}
	return float64(sum) / float64(len(v))
}
