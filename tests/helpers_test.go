package tests

import (
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/alparslanahmed/keyword-market-mcp/internal/cache"
	"github.com/alparslanahmed/keyword-market-mcp/internal/db"
	"github.com/alparslanahmed/keyword-market-mcp/internal/googleads"
	"github.com/alparslanahmed/keyword-market-mcp/internal/mcp"
	"github.com/alparslanahmed/keyword-market-mcp/internal/ratelimit"
	"github.com/alparslanahmed/keyword-market-mcp/internal/validation"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type testStack struct {
	store    *db.Store
	cache    *cache.Cache
	registry *googleads.TargetingRegistry
	service  *validation.Service
}

// newTestStack wires the full service with the mock adapter and a temp SQLite DB.
// A high rate limit keeps tests fast. No real credentials are involved.
func newTestStack(t *testing.T) testStack {
	t.Helper()
	logger := discardLogger()
	dsn := filepath.Join(t.TempDir(), "test.db")
	gdb, err := db.Open(dsn, logger)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	store := db.NewStore(gdb)
	registry, err := googleads.LoadTargetingRegistry()
	if err != nil {
		t.Fatalf("load targeting: %v", err)
	}
	c := cache.New(store, 7, 30)
	limiter := ratelimit.NewLimiter(1000)
	service := validation.New(googleads.NewMockGoogleAdsAdapter(), c, limiter, store, registry, logger, validation.Config{})
	return testStack{store: store, cache: c, registry: registry, service: service}
}

func newTestMCPServer(t *testing.T) (*mcp.Server, testStack) {
	t.Helper()
	st := newTestStack(t)
	server := mcp.NewServer(mcp.Deps{
		Name:     "keyword-market",
		Version:  "0.1.0",
		Logger:   discardLogger(),
		Service:  st.service,
		Store:    st.store,
		Cache:    st.cache,
		Registry: st.registry,
	})
	return server, st
}
