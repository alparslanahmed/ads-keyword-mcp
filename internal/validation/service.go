package validation

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/alparslanahmed/keyword-market-mcp/internal/cache"
	"github.com/alparslanahmed/keyword-market-mcp/internal/db"
	"github.com/alparslanahmed/keyword-market-mcp/internal/googleads"
	"github.com/alparslanahmed/keyword-market-mcp/internal/ratelimit"
	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// Service orchestrates keyword discovery, metrics, scoring, clustering, and
// product-idea validation. It is the single entry point used by the MCP tools,
// the HTTP debug server, and the CLI.
type Service struct {
	adapter  googleads.GoogleAdsAdapter
	cache    *cache.Cache
	limiter  *ratelimit.Limiter
	store    *db.Store
	registry *googleads.TargetingRegistry
	logger   *slog.Logger

	retryCfg        ratelimit.RetryConfig
	network         string
	defaultCountry  string
	defaultLanguage string
	defaultMaxIdeas int
}

// Config configures a Service.
type Config struct {
	Network                string
	DefaultCountry         string
	DefaultLanguage        string
	DefaultMaxKeywordIdeas int
	RetryConfig            ratelimit.RetryConfig
}

// New builds a Service, applying sensible defaults for any zero-valued config.
func New(adapter googleads.GoogleAdsAdapter, c *cache.Cache, limiter *ratelimit.Limiter, store *db.Store, registry *googleads.TargetingRegistry, logger *slog.Logger, cfg Config) *Service {
	if cfg.Network == "" {
		cfg.Network = googleads.NetworkGoogleSearch
	}
	if cfg.DefaultCountry == "" {
		cfg.DefaultCountry = "Turkey"
	}
	if cfg.DefaultLanguage == "" {
		cfg.DefaultLanguage = "Turkish"
	}
	if cfg.DefaultMaxKeywordIdeas <= 0 {
		cfg.DefaultMaxKeywordIdeas = 300
	}
	if cfg.RetryConfig.MaxRetries == 0 && cfg.RetryConfig.BaseDelay == 0 {
		cfg.RetryConfig = ratelimit.DefaultRetryConfig()
	}
	return &Service{
		adapter:         adapter,
		cache:           c,
		limiter:         limiter,
		store:           store,
		registry:        registry,
		logger:          logger,
		retryCfg:        cfg.RetryConfig,
		network:         cfg.Network,
		defaultCountry:  cfg.DefaultCountry,
		defaultLanguage: cfg.DefaultLanguage,
		defaultMaxIdeas: cfg.DefaultMaxKeywordIdeas,
	}
}

// DiscoverInput is the discover_keywords request.
type DiscoverInput struct {
	SeedKeywords []string
	URL          string
	Country      string
	Language     string
	Limit        int
	ForceRefresh bool
}

// MetricsInput is the get_keyword_metrics request.
type MetricsInput struct {
	Keywords     []string
	Country      string
	Language     string
	ForceRefresh bool
}

func (s *Service) resolveTargeting(country, language string) (googleads.Targeting, error) {
	if strings.TrimSpace(country) == "" {
		country = s.defaultCountry
	}
	if strings.TrimSpace(language) == "" {
		language = s.defaultLanguage
	}
	return s.registry.Resolve(country, language)
}

// DiscoverKeywords generates keyword ideas from seeds and/or a URL.
func (s *Service) DiscoverKeywords(ctx context.Context, in DiscoverInput) (*DiscoverResult, error) {
	t, err := s.resolveTargeting(in.Country, in.Language)
	if err != nil {
		return nil, err
	}
	seeds := utils.NormalizeKeywords(in.SeedKeywords)
	url := s.sanitizeURL(in.URL)
	if len(seeds) == 0 && url == "" {
		return nil, utils.NewValidationInputError("Provide at least one seed keyword or a valid URL.")
	}
	limit := in.Limit
	if limit <= 0 {
		limit = s.defaultMaxIdeas
	}

	ideas, cacheHit, err := s.fetchIdeas(ctx, t, seeds, url, limit, false, in.ForceRefresh)
	if err != nil {
		return nil, err
	}
	s.persistIdeas(t, ideas)

	out := &DiscoverResult{Keywords: make([]KeywordIdeaOut, 0, len(ideas))}
	for _, idea := range ideas {
		out.Keywords = append(out.Keywords, toIdeaOut(idea, cacheHit))
	}
	return out, nil
}

// GetKeywordMetrics fetches historical metrics for a list of keywords.
func (s *Service) GetKeywordMetrics(ctx context.Context, in MetricsInput) (*MetricsResult, error) {
	t, err := s.resolveTargeting(in.Country, in.Language)
	if err != nil {
		return nil, err
	}
	keywords := utils.NormalizeKeywords(in.Keywords)
	if len(keywords) == 0 {
		return nil, utils.NewValidationInputError("Provide at least one keyword.")
	}

	metrics, cacheHit, err := s.fetchHistorical(ctx, t, keywords, in.ForceRefresh)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	for _, m := range metrics {
		s.persistMetricSnapshot(t, m, now)
	}

	out := &MetricsResult{Metrics: make([]MetricOut, 0, len(metrics))}
	for _, m := range metrics {
		out.Metrics = append(out.Metrics, toMetricOut(m, now, cacheHit))
	}
	return out, nil
}

// sanitizeURL validates a URL, dropping it (with a warning) rather than failing
// the whole request if it is invalid.
func (s *Service) sanitizeURL(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	clean, err := utils.ValidateURL(raw)
	if err != nil {
		s.logger.Warn("ignoring invalid url seed")
		return ""
	}
	return clean
}

// fetchIdeas resolves keyword ideas through the cache, falling back to a live
// API call, with a URL-seed fallback and a stale-cache fallback on API failure.
func (s *Service) fetchIdeas(ctx context.Context, t googleads.Targeting, seeds []string, url string, limit int, includeAdult, forceRefresh bool) ([]googleads.KeywordIdea, bool, error) {
	key := cache.KeywordIdeasKey(seeds, url, t.Country, t.Language, s.network)
	var lk cache.Lookup
	if !forceRefresh {
		lk, _ = s.cache.Get(key)
		if lk.Found && lk.Fresh {
			if ideas, err := decodeIdeas(lk.ValueJSON); err == nil {
				return ideas, true, nil
			}
		}
	}

	input := googleads.GenerateKeywordIdeasInput{
		SeedKeywords:      seeds,
		URL:               url,
		GeoTargetConstant: t.GeoTargetConstant,
		LanguageConstant:  t.LanguageConstant,
		Network:           s.network,
		Limit:             limit,
		IncludeAdult:      includeAdult,
	}
	out, err := s.callIdeas(ctx, input)
	if err != nil && url != "" && len(seeds) > 0 && shouldFallbackWithoutURL(err) {
		s.logger.Warn("retrying keyword ideas without url seed")
		input.URL = ""
		out, err = s.callIdeas(ctx, input)
	}
	if err != nil {
		if lk.Found {
			if ideas, derr := decodeIdeas(lk.ValueJSON); derr == nil {
				s.logger.Warn("returning stale cached keyword ideas after api error")
				return ideas, true, nil
			}
		}
		return nil, false, err
	}

	if data, merr := json.Marshal(out.Ideas); merr == nil {
		if cerr := s.cache.Set(key, cache.TypeKeywordIdeas, string(data)); cerr != nil {
			s.logger.Warn("failed to cache keyword ideas")
		}
	}
	return out.Ideas, false, nil
}

func (s *Service) fetchHistorical(ctx context.Context, t googleads.Targeting, keywords []string, forceRefresh bool) ([]googleads.KeywordMetric, bool, error) {
	key := cache.HistoricalMetricsKey(keywords, t.Country, t.Language, s.network)
	var lk cache.Lookup
	if !forceRefresh {
		lk, _ = s.cache.Get(key)
		if lk.Found && lk.Fresh {
			if metrics, err := decodeMetrics(lk.ValueJSON); err == nil {
				return metrics, true, nil
			}
		}
	}

	input := googleads.GenerateHistoricalMetricsInput{
		Keywords:          keywords,
		GeoTargetConstant: t.GeoTargetConstant,
		LanguageConstant:  t.LanguageConstant,
		Network:           s.network,
	}
	out, err := s.callHistorical(ctx, input)
	if err != nil {
		if lk.Found {
			if metrics, derr := decodeMetrics(lk.ValueJSON); derr == nil {
				s.logger.Warn("returning stale cached metrics after api error")
				return metrics, true, nil
			}
		}
		return nil, false, err
	}

	if data, merr := json.Marshal(out.Metrics); merr == nil {
		if cerr := s.cache.Set(key, cache.TypeHistoricalMetrics, string(data)); cerr != nil {
			s.logger.Warn("failed to cache historical metrics")
		}
	}
	return out.Metrics, false, nil
}

func (s *Service) callIdeas(ctx context.Context, input googleads.GenerateKeywordIdeasInput) (*googleads.GenerateKeywordIdeasOutput, error) {
	reqHash := utils.SHA256Hex("GenerateKeywordIdeas", input.GeoTargetConstant, input.LanguageConstant,
		strings.Join(utils.NormalizeKeywords(input.SeedKeywords), "|"), input.URL, input.Network)
	start := time.Now()
	if err := s.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	var out *googleads.GenerateKeywordIdeasOutput
	err := ratelimit.Retry(ctx, s.retryCfg, func() error {
		var e error
		out, e = s.adapter.GenerateKeywordIdeas(ctx, input)
		return e
	}, nil)
	s.logApiCall("KeywordPlanIdeaService", "GenerateKeywordIdeas", reqHash, start, err)
	return out, err
}

func (s *Service) callHistorical(ctx context.Context, input googleads.GenerateHistoricalMetricsInput) (*googleads.GenerateHistoricalMetricsOutput, error) {
	reqHash := utils.SHA256Hex("GenerateHistoricalMetrics", input.GeoTargetConstant, input.LanguageConstant,
		strings.Join(utils.NormalizeKeywords(input.Keywords), "|"), input.Network)
	start := time.Now()
	if err := s.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	var out *googleads.GenerateHistoricalMetricsOutput
	err := ratelimit.Retry(ctx, s.retryCfg, func() error {
		var e error
		out, e = s.adapter.GenerateHistoricalMetrics(ctx, input)
		return e
	}, nil)
	s.logApiCall("KeywordPlanIdeaService", "GenerateKeywordHistoricalMetrics", reqHash, start, err)
	return out, err
}

func (s *Service) logApiCall(service, method, reqHash string, start time.Time, err error) {
	fin := time.Now()
	entry := &db.ApiCallLog{
		Service:     service,
		Method:      method,
		RequestHash: reqHash,
		StartedAt:   start,
		FinishedAt:  fin,
		DurationMs:  fin.Sub(start).Milliseconds(),
		Status:      "ok",
	}
	if err != nil {
		ae := utils.AsAppError(err)
		entry.Status = "error"
		entry.ErrorCode = string(ae.Code)
		entry.ErrorMessage = ae.Message // safe message only; never the raw cause
	}
	if logErr := s.store.LogApiCall(entry); logErr != nil {
		s.logger.Warn("failed to write api call log")
	}
}

func (s *Service) persistIdeas(t googleads.Targeting, ideas []googleads.KeywordIdea) {
	now := time.Now()
	for _, idea := range ideas {
		s.persistMetricSnapshot(t, idea.KeywordMetric, now)
	}
}

// persistMetricSnapshot upserts the keyword and stores a metric snapshot. Errors
// are logged but do not fail the caller, since persistence is observational.
func (s *Service) persistMetricSnapshot(t googleads.Targeting, m googleads.KeywordMetric, now time.Time) (string, string) {
	norm := utils.NormalizeKeyword(m.Keyword)
	if norm == "" {
		return "", ""
	}
	kw, err := s.store.UpsertKeyword(m.Keyword, norm)
	if err != nil {
		s.logger.Warn("failed to upsert keyword")
		return "", ""
	}
	mvJSON, _ := json.Marshal(m.MonthlySearchVolumes)
	rawJSON := ""
	if m.RawPayload != nil {
		if b, mErr := json.Marshal(m.RawPayload); mErr == nil {
			rawJSON = string(b)
		}
	}
	row := &db.KeywordMetric{
		KeywordID:                kw.ID,
		Country:                  t.Country,
		Language:                 t.Language,
		Network:                  s.network,
		AvgMonthlySearches:       m.AvgMonthlySearches,
		Competition:              googleads.NormalizeCompetition(m.Competition),
		CompetitionIndex:         m.CompetitionIndex,
		LowTopOfPageBidMicros:    m.LowTopOfPageBidMicros,
		HighTopOfPageBidMicros:   m.HighTopOfPageBidMicros,
		MonthlySearchVolumesJSON: string(mvJSON),
		Source:                   m.Source,
		RawPayloadJSON:           rawJSON,
		FetchedAt:                now,
		CreatedAt:                now,
	}
	if err := s.store.CreateMetric(row); err != nil {
		s.logger.Warn("failed to store keyword metric")
		return kw.ID, ""
	}
	return kw.ID, row.ID
}

func shouldFallbackWithoutURL(err error) bool {
	code := utils.AsAppError(err).Code
	return code == utils.CodeNoKeywordData || code == utils.CodeGoogleAdsAPI
}

func decodeIdeas(value string) ([]googleads.KeywordIdea, error) {
	var ideas []googleads.KeywordIdea
	if err := json.Unmarshal([]byte(value), &ideas); err != nil {
		return nil, err
	}
	return ideas, nil
}

func decodeMetrics(value string) ([]googleads.KeywordMetric, error) {
	var metrics []googleads.KeywordMetric
	if err := json.Unmarshal([]byte(value), &metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}
