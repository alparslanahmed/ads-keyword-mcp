package db

import "time"

// ResearchRun is one product-idea validation run.
type ResearchRun struct {
	ID                    string `gorm:"primaryKey"`
	Idea                  string
	Country               string
	Language              string
	SeedInputJSON         string
	Status                string
	Verdict               string
	DemandScore           float64
	CompetitionScore      float64
	CommercialIntentScore float64
	OpportunityScore      float64
	SummaryJSON           string
	CreatedAt             time.Time
	UpdatedAt             time.Time
	CompletedAt           *time.Time
	ErrorMessage          string
}

// Keyword is a deduplicated keyword string.
type Keyword struct {
	ID             string `gorm:"primaryKey"`
	Text           string
	NormalizedText string `gorm:"uniqueIndex"`
	CreatedAt      time.Time
}

// KeywordMetric is a stored metric snapshot for a keyword/targeting combination.
type KeywordMetric struct {
	ID                       string `gorm:"primaryKey"`
	KeywordID                string `gorm:"index"`
	Country                  string
	Language                 string
	Network                  string
	AvgMonthlySearches       *int64
	Competition              string
	CompetitionIndex         *float64
	LowTopOfPageBidMicros    *int64
	HighTopOfPageBidMicros   *int64
	MonthlySearchVolumesJSON string
	Source                   string
	RawPayloadJSON           string
	FetchedAt                time.Time
	CreatedAt                time.Time
}

// ResearchRunKeyword links a run to a keyword with the scoring snapshot.
type ResearchRunKeyword struct {
	ID                    string `gorm:"primaryKey"`
	ResearchRunID         string `gorm:"index"`
	KeywordID             string `gorm:"index"`
	MetricID              string
	SourceSeed            string
	OpportunityScore      float64
	VolumeScore           float64
	CompetitionScore      float64
	CommercialIntentScore float64
	Intent                string
	ClusterName           string
	NotesJSON             string
	CreatedAt             time.Time
}

// ApiCallLog records every Google Ads API call for observability.
type ApiCallLog struct {
	ID           string `gorm:"primaryKey"`
	Service      string
	Method       string
	RequestHash  string
	Status       string
	ErrorCode    string
	ErrorMessage string
	StartedAt    time.Time
	FinishedAt   time.Time
	DurationMs   int64
}

// CacheEntry is a generic key/value cache row with TTL.
type CacheEntry struct {
	ID        string `gorm:"primaryKey"`
	Key       string `gorm:"uniqueIndex"`
	Type      string
	ValueJSON string
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AllModels lists every model for AutoMigrate.
func AllModels() []any {
	return []any{
		&ResearchRun{},
		&Keyword{},
		&KeywordMetric{},
		&ResearchRunKeyword{},
		&ApiCallLog{},
		&CacheEntry{},
	}
}
