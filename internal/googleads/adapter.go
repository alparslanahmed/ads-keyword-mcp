package googleads

import (
	"context"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// GoogleAdsAdapter is the clean boundary between the application and Google Ads.
// Implementations: MockGoogleAdsAdapter (local data, no credentials) and
// RealGoogleAdsAdapter (live Google Ads REST API).
type GoogleAdsAdapter interface {
	GenerateKeywordIdeas(ctx context.Context, input GenerateKeywordIdeasInput) (*GenerateKeywordIdeasOutput, error)
	GenerateHistoricalMetrics(ctx context.Context, input GenerateHistoricalMetricsInput) (*GenerateHistoricalMetricsOutput, error)
}

// unconfiguredAdapter is used when no credentials are configured and mock mode is
// off. It lets the server start and list tools while returning a safe
// MISSING_CREDENTIALS error on any data call.
type unconfiguredAdapter struct{}

// NewUnconfiguredAdapter returns an adapter that reports missing credentials.
func NewUnconfiguredAdapter() GoogleAdsAdapter { return unconfiguredAdapter{} }

func (unconfiguredAdapter) GenerateKeywordIdeas(context.Context, GenerateKeywordIdeasInput) (*GenerateKeywordIdeasOutput, error) {
	return nil, utils.NewMissingCredentialsError(nil)
}

func (unconfiguredAdapter) GenerateHistoricalMetrics(context.Context, GenerateHistoricalMetricsInput) (*GenerateHistoricalMetricsOutput, error) {
	return nil, utils.NewMissingCredentialsError(nil)
}
