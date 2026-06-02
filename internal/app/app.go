package app

import (
	"log/slog"

	"gorm.io/gorm"

	"github.com/alparslanahmed/keyword-market-mcp/internal/cache"
	"github.com/alparslanahmed/keyword-market-mcp/internal/db"
	"github.com/alparslanahmed/keyword-market-mcp/internal/googleads"
	"github.com/alparslanahmed/keyword-market-mcp/internal/ratelimit"
	"github.com/alparslanahmed/keyword-market-mcp/internal/validation"
)

// App is the wired application: configuration plus all shared dependencies. The
// MCP server, HTTP server, and CLI all build on the same App.
type App struct {
	Config   *Config
	Logger   *slog.Logger
	DB       *gorm.DB
	Store    *db.Store
	Adapter  googleads.GoogleAdsAdapter
	Cache    *cache.Cache
	Limiter  *ratelimit.Limiter
	Registry *googleads.TargetingRegistry
	Service  *validation.Service
}

// New builds the application from configuration.
func New(cfg *Config, logger *slog.Logger) (*App, error) {
	gdb, err := db.Open(cfg.DatabaseDSN, logger)
	if err != nil {
		return nil, err
	}
	store := db.NewStore(gdb)

	registry, err := googleads.LoadTargetingRegistry()
	if err != nil {
		return nil, err
	}

	adapter := selectAdapter(cfg, logger)
	c := cache.New(store, cfg.CacheKeywordIdeasDays, cfg.CacheHistoricalMetricsDays)
	limiter := ratelimit.NewLimiter(cfg.RateLimitRPS)

	service := validation.New(adapter, c, limiter, store, registry, logger, validation.Config{
		RetryConfig: ratelimit.DefaultRetryConfig(),
	})

	return &App{
		Config:   cfg,
		Logger:   logger,
		DB:       gdb,
		Store:    store,
		Adapter:  adapter,
		Cache:    c,
		Limiter:  limiter,
		Registry: registry,
		Service:  service,
	}, nil
}

func selectAdapter(cfg *Config, logger *slog.Logger) googleads.GoogleAdsAdapter {
	if cfg.UseMock {
		logger.Info("using mock Google Ads adapter (no credentials required)")
		return googleads.NewMockGoogleAdsAdapter()
	}
	real, err := googleads.NewRealGoogleAdsAdapter(cfg.GoogleAds, logger)
	if err != nil {
		logger.Warn("Google Ads credentials are incomplete; tool calls will return MISSING_CREDENTIALS until configured (or set GOOGLE_ADS_MOCK=true)")
		return googleads.NewUnconfiguredAdapter()
	}
	logger.Info("using real Google Ads adapter")
	return real
}

// Close releases resources held by the App.
func (a *App) Close() error {
	return db.Close(a.DB)
}
