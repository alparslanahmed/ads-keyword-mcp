package googleads

// historicalRequest is the GenerateKeywordHistoricalMetrics REST request body.
type historicalRequest struct {
	Keywords           []string `json:"keywords"`
	Language           string   `json:"language,omitempty"`
	GeoTargetConstants []string `json:"geoTargetConstants,omitempty"`
	KeywordPlanNetwork string   `json:"keywordPlanNetwork,omitempty"`
}

type historicalResponse struct {
	Results []struct {
		Text           string     `json:"text"`
		KeywordMetrics apiMetrics `json:"keywordMetrics"`
	} `json:"results"`
}

func buildHistoricalRequest(input GenerateHistoricalMetricsInput) historicalRequest {
	return historicalRequest{
		Keywords:           input.Keywords,
		Language:           input.LanguageConstant,
		GeoTargetConstants: []string{input.GeoTargetConstant},
		KeywordPlanNetwork: networkOrDefault(input.Network),
	}
}

func (r historicalResponse) normalize(source string) []KeywordMetric {
	out := make([]KeywordMetric, 0, len(r.Results))
	for _, res := range r.Results {
		out = append(out, normalizeAPIMetrics(res.Text, res.KeywordMetrics, source, res))
	}
	return out
}
