package app

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

	"github.com/alparslanahmed/keyword-market-mcp/internal/googleads"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	GoogleAds googleads.Credentials
	UseMock   bool

	DatabaseDSN string
	LogLevel    string
	AppEnv      string

	RateLimitRPS               float64
	CacheKeywordIdeasDays      int
	CacheHistoricalMetricsDays int

	MCPServerName    string
	MCPServerVersion string

	HTTPPort    string
	DebugAPIKey string
}

// Load reads configuration from the environment, loading a .env file first if
// present. Secrets are read into Config but never logged or serialized.
func Load() *Config {
	// Best-effort: a missing .env is fine (env vars may be set directly).
	_ = godotenv.Load()

	appEnv := getEnv("APP_ENV", "development")
	useMock := appEnv == "test" || isTruthy(os.Getenv("GOOGLE_ADS_MOCK"))

	return &Config{
		GoogleAds: googleads.Credentials{
			DeveloperToken:  os.Getenv("GOOGLE_ADS_DEVELOPER_TOKEN"),
			ClientID:        os.Getenv("GOOGLE_ADS_CLIENT_ID"),
			ClientSecret:    os.Getenv("GOOGLE_ADS_CLIENT_SECRET"),
			RefreshToken:    os.Getenv("GOOGLE_ADS_REFRESH_TOKEN"),
			LoginCustomerID: os.Getenv("GOOGLE_ADS_LOGIN_CUSTOMER_ID"),
			CustomerID:      os.Getenv("GOOGLE_ADS_CUSTOMER_ID"),
			APIVersion:      os.Getenv("GOOGLE_ADS_API_VERSION"),
		},
		UseMock:                    useMock,
		DatabaseDSN:                getEnv("DATABASE_DSN", "keyword_market.db"),
		LogLevel:                   getEnv("LOG_LEVEL", "info"),
		AppEnv:                     appEnv,
		RateLimitRPS:               getEnvFloat("KEYWORD_PLANNING_RATE_LIMIT_RPS", 1),
		CacheKeywordIdeasDays:      getEnvInt("CACHE_KEYWORD_IDEAS_DAYS", 7),
		CacheHistoricalMetricsDays: getEnvInt("CACHE_HISTORICAL_METRICS_DAYS", 30),
		MCPServerName:              getEnv("MCP_SERVER_NAME", "keyword-market"),
		MCPServerVersion:           getEnv("MCP_SERVER_VERSION", "0.1.0"),
		HTTPPort:                   getEnv("HTTP_PORT", "8080"),
		DebugAPIKey:                os.Getenv("DEBUG_API_KEY"),
	}
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func isTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
