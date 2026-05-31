package googleads

// Network targeting values accepted by the Keyword Planner.
const (
	NetworkGoogleSearch            = "GOOGLE_SEARCH"
	NetworkGoogleSearchAndPartners = "GOOGLE_SEARCH_AND_PARTNERS"
)

// Competition levels as normalized by this server.
const (
	CompetitionLow     = "LOW"
	CompetitionMedium  = "MEDIUM"
	CompetitionHigh    = "HIGH"
	CompetitionUnknown = "UNKNOWN"
)

// Source values record where a metric came from.
const (
	SourceMock      = "mock"
	SourceGoogleAds = "google_ads"
)

// MonthlySearchVolume is a single month data point from historical metrics.
type MonthlySearchVolume struct {
	Year            int   `json:"year"`
	Month           int   `json:"month"`
	MonthlySearches int64 `json:"monthlySearches"`
}

// KeywordMetric is the internal normalized representation of keyword data,
// shared by both keyword-idea and historical-metric responses. Bid values are
// stored in micros; conversion to display units happens at the output edge.
type KeywordMetric struct {
	Keyword                string                `json:"keyword"`
	AvgMonthlySearches     *int64                `json:"avgMonthlySearches"`
	MonthlySearchVolumes   []MonthlySearchVolume `json:"monthlySearchVolumes"`
	Competition            string                `json:"competition"`
	CompetitionIndex       *float64              `json:"competitionIndex"`
	LowTopOfPageBidMicros  *int64                `json:"lowTopOfPageBidMicros"`
	HighTopOfPageBidMicros *int64                `json:"highTopOfPageBidMicros"`
	Source                 string                `json:"source"`
	RawPayload             any                   `json:"-"`
}

// KeywordIdea extends a metric with the seed that produced it.
type KeywordIdea struct {
	KeywordMetric
	SourceSeed string `json:"sourceSeed"`
}

// GenerateKeywordIdeasInput is the adapter input for idea generation.
type GenerateKeywordIdeasInput struct {
	SeedKeywords      []string
	URL               string
	GeoTargetConstant string
	LanguageConstant  string
	Network           string
	Limit             int
	IncludeAdult      bool
	CustomerID        string
}

// GenerateKeywordIdeasOutput is the adapter output for idea generation.
type GenerateKeywordIdeasOutput struct {
	Ideas []KeywordIdea
}

// GenerateHistoricalMetricsInput is the adapter input for historical metrics.
type GenerateHistoricalMetricsInput struct {
	Keywords          []string
	GeoTargetConstant string
	LanguageConstant  string
	Network           string
	CustomerID        string
}

// GenerateHistoricalMetricsOutput is the adapter output for historical metrics.
type GenerateHistoricalMetricsOutput struct {
	Metrics []KeywordMetric
}

// NormalizeCompetition maps an arbitrary competition string to a known level.
func NormalizeCompetition(s string) string {
	switch s {
	case CompetitionLow, CompetitionMedium, CompetitionHigh:
		return s
	case "":
		return CompetitionUnknown
	default:
		return CompetitionUnknown
	}
}
