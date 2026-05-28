package googleads

import "strconv"

// apiMetrics mirrors the Google Ads REST metrics object. The REST transport
// encodes int64 fields as JSON strings, so they are parsed leniently.
type apiMetrics struct {
	Competition            string       `json:"competition"`
	AvgMonthlySearches     string       `json:"avgMonthlySearches"`
	CompetitionIndex       string       `json:"competitionIndex"`
	LowTopOfPageBidMicros  string       `json:"lowTopOfPageBidMicros"`
	HighTopOfPageBidMicros string       `json:"highTopOfPageBidMicros"`
	MonthlySearchVolumes   []apiMonthly `json:"monthlySearchVolumes"`
}

type apiMonthly struct {
	Year            string `json:"year"`
	Month           string `json:"month"`
	MonthlySearches string `json:"monthlySearches"`
}

// ideasRequest is the GenerateKeywordIdeas REST request body.
type ideasRequest struct {
	Language             string          `json:"language,omitempty"`
	GeoTargetConstants   []string        `json:"geoTargetConstants,omitempty"`
	KeywordPlanNetwork   string          `json:"keywordPlanNetwork,omitempty"`
	IncludeAdultKeywords bool            `json:"includeAdultKeywords"`
	PageSize             int             `json:"pageSize,omitempty"`
	KeywordSeed          *keywordSeed    `json:"keywordSeed,omitempty"`
	URLSeed              *urlSeed        `json:"urlSeed,omitempty"`
	KeywordAndURLSeed    *keywordURLSeed `json:"keywordAndUrlSeed,omitempty"`
}

type keywordSeed struct {
	Keywords []string `json:"keywords"`
}
type urlSeed struct {
	URL string `json:"url"`
}
type keywordURLSeed struct {
	URL      string   `json:"url"`
	Keywords []string `json:"keywords"`
}

type ideasResponse struct {
	Results []struct {
		Text               string     `json:"text"`
		KeywordIdeaMetrics apiMetrics `json:"keywordIdeaMetrics"`
	} `json:"results"`
	NextPageToken string `json:"nextPageToken"`
}

func buildIdeasRequest(input GenerateKeywordIdeasInput) ideasRequest {
	req := ideasRequest{
		Language:             input.LanguageConstant,
		GeoTargetConstants:   []string{input.GeoTargetConstant},
		KeywordPlanNetwork:   networkOrDefault(input.Network),
		IncludeAdultKeywords: input.IncludeAdult,
		PageSize:             input.Limit,
	}
	hasKeywords := len(input.SeedKeywords) > 0
	hasURL := input.URL != ""
	switch {
	case hasKeywords && hasURL:
		req.KeywordAndURLSeed = &keywordURLSeed{URL: input.URL, Keywords: input.SeedKeywords}
	case hasURL:
		req.URLSeed = &urlSeed{URL: input.URL}
	default:
		req.KeywordSeed = &keywordSeed{Keywords: input.SeedKeywords}
	}
	return req
}

func (r ideasResponse) normalize(source string) []KeywordIdea {
	out := make([]KeywordIdea, 0, len(r.Results))
	for _, res := range r.Results {
		out = append(out, KeywordIdea{
			KeywordMetric: normalizeAPIMetrics(res.Text, res.KeywordIdeaMetrics, source, res),
		})
	}
	return out
}

func normalizeAPIMetrics(text string, m apiMetrics, source string, raw any) KeywordMetric {
	return KeywordMetric{
		Keyword:                text,
		AvgMonthlySearches:     parseInt64Ptr(m.AvgMonthlySearches),
		CompetitionIndex:       parseFloatPtr(m.CompetitionIndex),
		Competition:            NormalizeCompetition(m.Competition),
		LowTopOfPageBidMicros:  parseInt64Ptr(m.LowTopOfPageBidMicros),
		HighTopOfPageBidMicros: parseInt64Ptr(m.HighTopOfPageBidMicros),
		MonthlySearchVolumes:   normalizeMonthly(m.MonthlySearchVolumes),
		Source:                 source,
		RawPayload:             raw,
	}
}

func normalizeMonthly(in []apiMonthly) []MonthlySearchVolume {
	out := make([]MonthlySearchVolume, 0, len(in))
	for _, mv := range in {
		year, _ := strconv.Atoi(mv.Year)
		searches, _ := strconv.ParseInt(mv.MonthlySearches, 10, 64)
		out = append(out, MonthlySearchVolume{
			Year:            year,
			Month:           monthToInt(mv.Month),
			MonthlySearches: searches,
		})
	}
	return out
}

func networkOrDefault(n string) string {
	if n == "" {
		return NetworkGoogleSearch
	}
	return n
}

func parseInt64Ptr(s string) *int64 {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &v
}

func parseFloatPtr(s string) *float64 {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

// monthToInt maps the Google Ads month enum to 1-12. Numeric strings are also
// accepted for resilience.
func monthToInt(month string) int {
	switch month {
	case "JANUARY":
		return 1
	case "FEBRUARY":
		return 2
	case "MARCH":
		return 3
	case "APRIL":
		return 4
	case "MAY":
		return 5
	case "JUNE":
		return 6
	case "JULY":
		return 7
	case "AUGUST":
		return 8
	case "SEPTEMBER":
		return 9
	case "OCTOBER":
		return 10
	case "NOVEMBER":
		return 11
	case "DECEMBER":
		return 12
	default:
		if v, err := strconv.Atoi(month); err == nil {
			return v
		}
		return 0
	}
}
