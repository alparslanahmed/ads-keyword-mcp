package validation

// Verdict and confidence labels.
const (
	VerdictStrong       = "strong"
	VerdictPromising    = "promising"
	VerdictWeak         = "weak"
	VerdictInconclusive = "inconclusive"

	ConfidenceLow    = "low"
	ConfidenceMedium = "medium"
	ConfidenceHigh   = "high"
)

// VerdictSignals are the aggregate inputs to the verdict decision. Keeping this
// a plain struct makes the verdict logic pure and unit-testable.
type VerdictSignals struct {
	KeywordCount              int
	KeywordsWithVolume        int
	CommercialKeywords        int
	LowMedCompetitionKeywords int
	ClusterCount              int
	DemandScore               float64
	CommercialIntentScore     float64
	CompetitionScore          float64
	OpportunityScore          float64
}

// DetermineVerdict applies careful, factual thresholds. Insufficient data yields
// "inconclusive" rather than a false negative.
func DetermineVerdict(s VerdictSignals) string {
	if s.KeywordCount < 8 || s.KeywordsWithVolume == 0 {
		return VerdictInconclusive
	}
	switch {
	case s.DemandScore >= 60 &&
		s.CommercialKeywords >= 3 &&
		s.LowMedCompetitionKeywords >= 5 &&
		s.ClusterCount >= 3:
		return VerdictStrong
	case s.DemandScore >= 35 &&
		s.CommercialKeywords >= 1 &&
		s.LowMedCompetitionKeywords >= 2:
		return VerdictPromising
	default:
		return VerdictWeak
	}
}

// DetermineConfidence reflects how much data backs the verdict.
func DetermineConfidence(s VerdictSignals) string {
	switch {
	case s.KeywordCount >= 50 && s.KeywordsWithVolume >= 20:
		return ConfidenceHigh
	case s.KeywordCount >= 15 && s.KeywordsWithVolume >= 5:
		return ConfidenceMedium
	default:
		return ConfidenceLow
	}
}
