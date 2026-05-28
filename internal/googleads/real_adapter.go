package googleads

import (
	"context"
	"encoding/json"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// RealGoogleAdsAdapter calls the live Google Ads REST API. Credentials stay
// server-side and all errors are mapped to safe AppErrors before returning.
type RealGoogleAdsAdapter struct {
	client *httpClient
}

// NewRealGoogleAdsAdapter validates credentials and builds the adapter.
func NewRealGoogleAdsAdapter(creds Credentials) (*RealGoogleAdsAdapter, error) {
	if err := creds.validate(); err != nil {
		return nil, err
	}
	return &RealGoogleAdsAdapter{client: newHTTPClient(creds)}, nil
}

// GenerateKeywordIdeas calls KeywordPlanIdeaService.GenerateKeywordIdeas.
func (a *RealGoogleAdsAdapter) GenerateKeywordIdeas(ctx context.Context, input GenerateKeywordIdeasInput) (*GenerateKeywordIdeasOutput, error) {
	body, err := a.client.post(ctx, "generateKeywordIdeas", buildIdeasRequest(input))
	if err != nil {
		return nil, err
	}
	var parsed ideasResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, utils.NewGoogleAdsAPIError("Failed to parse the Google Ads response.", false, err)
	}
	ideas := parsed.normalize(SourceGoogleAds)
	if len(ideas) == 0 {
		return nil, utils.NewNoKeywordDataError("Google Ads returned no keyword ideas for this input.")
	}
	if input.Limit > 0 && len(ideas) > input.Limit {
		ideas = ideas[:input.Limit]
	}
	return &GenerateKeywordIdeasOutput{Ideas: ideas}, nil
}

// GenerateHistoricalMetrics calls KeywordPlanIdeaService.GenerateKeywordHistoricalMetrics.
func (a *RealGoogleAdsAdapter) GenerateHistoricalMetrics(ctx context.Context, input GenerateHistoricalMetricsInput) (*GenerateHistoricalMetricsOutput, error) {
	body, err := a.client.post(ctx, "generateKeywordHistoricalMetrics", buildHistoricalRequest(input))
	if err != nil {
		return nil, err
	}
	var parsed historicalResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, utils.NewGoogleAdsAPIError("Failed to parse the Google Ads response.", false, err)
	}
	metrics := parsed.normalize(SourceGoogleAds)
	if len(metrics) == 0 {
		return nil, utils.NewNoKeywordDataError("Google Ads returned no historical metrics for these keywords.")
	}
	return &GenerateHistoricalMetricsOutput{Metrics: metrics}, nil
}
