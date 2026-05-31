package clustering

import (
	"sort"

	"github.com/alparslanahmed/keyword-market-mcp/internal/intent"
	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// KeywordForCluster is the clustering input. Metric fields are optional; nil
// values are simply skipped in aggregates.
type KeywordForCluster struct {
	Keyword            string
	AvgMonthlySearches *int64
	CompetitionIndex   *float64
	OpportunityScore   *float64
	Intent             string
}

// Cluster is an aggregated group of related keywords.
type Cluster struct {
	Name                    string   `json:"name"`
	KeywordCount            int      `json:"keywordCount"`
	TotalSearchVolume       int64    `json:"totalSearchVolume"`
	AverageOpportunityScore float64  `json:"averageOpportunityScore"`
	MedianCompetitionIndex  *float64 `json:"medianCompetitionIndex"`
	DominantIntent          string   `json:"dominantIntent"`
	TopKeywords             []string `json:"topKeywords"`
}

const otherClusterName = "other"

// ClusterKeywords groups keywords by their most-shared meaningful token. The
// design is deliberately simple (token frequency, no embeddings) but isolated
// behind this function so an embedding-based implementation can replace it
// later without changing callers.
func ClusterKeywords(items []KeywordForCluster, language string) []Cluster {
	clusters, _ := ClusterWithAssignments(items, language)
	return clusters
}

// ClusterWithAssignments returns the clusters and, aligned to the input order,
// the cluster name assigned to each keyword. Callers that need per-keyword
// cluster labels (e.g. to annotate scored keywords) use this form.
func ClusterWithAssignments(items []KeywordForCluster, language string) ([]Cluster, []string) {
	if len(items) == 0 {
		return []Cluster{}, nil
	}

	tokensByItem := make([][]string, len(items))
	globalFreq := make(map[string]int)
	for i, it := range items {
		toks := meaningfulTokens(it.Keyword)
		tokensByItem[i] = toks
		seen := make(map[string]struct{}, len(toks))
		for _, t := range toks {
			if _, ok := seen[t]; ok {
				continue
			}
			seen[t] = struct{}{}
			globalFreq[t]++
		}
	}

	groups := make(map[string][]int)
	var order []string
	assignment := make([]string, len(items))
	for i := range items {
		key := chooseClusterToken(tokensByItem[i], globalFreq)
		assignment[i] = key
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], i)
	}

	clusters := make([]Cluster, 0, len(groups))
	for _, key := range order {
		clusters = append(clusters, buildCluster(key, groups[key], items, language))
	}

	// Largest, highest-volume clusters first for stable, useful ordering.
	sort.SliceStable(clusters, func(i, j int) bool {
		if clusters[i].TotalSearchVolume != clusters[j].TotalSearchVolume {
			return clusters[i].TotalSearchVolume > clusters[j].TotalSearchVolume
		}
		if clusters[i].KeywordCount != clusters[j].KeywordCount {
			return clusters[i].KeywordCount > clusters[j].KeywordCount
		}
		return clusters[i].Name < clusters[j].Name
	})
	return clusters, assignment
}

func meaningfulTokens(keyword string) []string {
	tokens := utils.Tokenize(keyword)
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if IsStopword(t) {
			continue
		}
		if len([]rune(t)) < 2 {
			continue
		}
		out = append(out, t)
	}
	return out
}

// chooseClusterToken groups a keyword by its most distinctive SHARED token: the
// meaningful token with the lowest document frequency that still appears in at
// least two keywords. This avoids collapsing everything into one cluster keyed
// on the ubiquitous head noun (e.g. "espresso"), instead surfacing themed
// modifier clusters ("price", "review", "cheap", ...). Ties break by the longer,
// then alphabetically-earlier token for determinism. Keywords whose tokens are
// all unique are named by their longest token; keywords with no meaningful
// tokens fall into the "other" bucket.
func chooseClusterToken(tokens []string, globalFreq map[string]int) string {
	if len(tokens) == 0 {
		return otherClusterName
	}
	best := ""
	bestFreq := 0
	for _, t := range tokens {
		f := globalFreq[t]
		if f < 2 {
			continue
		}
		if best == "" || f < bestFreq ||
			(f == bestFreq && len(t) > len(best)) ||
			(f == bestFreq && len(t) == len(best) && t < best) {
			best, bestFreq = t, f
		}
	}
	if best != "" {
		return best
	}
	longest := tokens[0]
	for _, t := range tokens[1:] {
		if len(t) > len(longest) || (len(t) == len(longest) && t < longest) {
			longest = t
		}
	}
	return longest
}

func buildCluster(name string, idxs []int, items []KeywordForCluster, language string) Cluster {
	c := Cluster{Name: name, KeywordCount: len(idxs)}

	var oppSum float64
	var oppCount int
	var compValues []float64
	intentCounts := make(map[string]int)

	type ranked struct {
		keyword string
		volume  int64
		opp     float64
	}
	rankable := make([]ranked, 0, len(idxs))

	for _, i := range idxs {
		it := items[i]
		var vol int64
		if it.AvgMonthlySearches != nil {
			vol = *it.AvgMonthlySearches
			c.TotalSearchVolume += vol
		}
		if it.OpportunityScore != nil {
			oppSum += *it.OpportunityScore
			oppCount++
		}
		if it.CompetitionIndex != nil {
			compValues = append(compValues, *it.CompetitionIndex)
		}
		kwIntent := it.Intent
		if kwIntent == "" {
			kwIntent = intent.ClassifyIntent(it.Keyword, language)
		}
		intentCounts[kwIntent]++
		opp := 0.0
		if it.OpportunityScore != nil {
			opp = *it.OpportunityScore
		}
		rankable = append(rankable, ranked{keyword: it.Keyword, volume: vol, opp: opp})
	}

	if oppCount > 0 {
		c.AverageOpportunityScore = round1(oppSum / float64(oppCount))
	}
	c.MedianCompetitionIndex = median(compValues)
	c.DominantIntent = dominant(intentCounts)

	sort.SliceStable(rankable, func(i, j int) bool {
		if rankable[i].opp != rankable[j].opp {
			return rankable[i].opp > rankable[j].opp
		}
		if rankable[i].volume != rankable[j].volume {
			return rankable[i].volume > rankable[j].volume
		}
		return rankable[i].keyword < rankable[j].keyword
	})
	for i, r := range rankable {
		if i >= 5 {
			break
		}
		c.TopKeywords = append(c.TopKeywords, r.keyword)
	}
	return c
}

func median(values []float64) *float64 {
	if len(values) == 0 {
		return nil
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	n := len(sorted)
	var m float64
	if n%2 == 1 {
		m = sorted[n/2]
	} else {
		m = (sorted[n/2-1] + sorted[n/2]) / 2
	}
	m = round1(m)
	return &m
}

func dominant(counts map[string]int) string {
	best := intent.Unknown
	bestCount := -1
	for k, v := range counts {
		if v > bestCount || (v == bestCount && k < best) {
			best, bestCount = k, v
		}
	}
	return best
}

func round1(v float64) float64 {
	return float64(int64(v*10+0.5)) / 10
}
