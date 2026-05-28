package cache

import (
	"sync/atomic"
	"time"

	"github.com/alparslanahmed/keyword-market-mcp/internal/db"
)

// Cache is a SQLite-backed TTL cache with in-process hit/miss counters.
type Cache struct {
	store      *db.Store
	ideasTTL   time.Duration
	metricsTTL time.Duration

	hits   atomic.Int64
	misses atomic.Int64
}

// New constructs a Cache. TTLs are given in days.
func New(store *db.Store, ideasTTLDays, metricsTTLDays int) *Cache {
	if ideasTTLDays <= 0 {
		ideasTTLDays = 7
	}
	if metricsTTLDays <= 0 {
		metricsTTLDays = 30
	}
	return &Cache{
		store:      store,
		ideasTTL:   time.Duration(ideasTTLDays) * 24 * time.Hour,
		metricsTTL: time.Duration(metricsTTLDays) * 24 * time.Hour,
	}
}

// TTLFor returns the configured TTL for a cache type.
func (c *Cache) TTLFor(cacheType string) time.Duration {
	if cacheType == TypeHistoricalMetrics {
		return c.metricsTTL
	}
	return c.ideasTTL
}

// Lookup is the result of a cache read.
type Lookup struct {
	Found     bool
	Fresh     bool
	ValueJSON string
	ExpiresAt time.Time
}

// Get reads a key and records a hit (found and fresh) or a miss otherwise. A
// stale-but-present entry is returned with Fresh=false so callers can fall back
// to it when a live API call fails.
func (c *Cache) Get(key string) (Lookup, error) {
	entry, err := c.store.GetCacheEntry(key)
	if err != nil {
		return Lookup{}, err
	}
	if entry == nil {
		c.misses.Add(1)
		return Lookup{Found: false}, nil
	}
	fresh := time.Now().Before(entry.ExpiresAt)
	if fresh {
		c.hits.Add(1)
	} else {
		c.misses.Add(1)
	}
	return Lookup{Found: true, Fresh: fresh, ValueJSON: entry.ValueJSON, ExpiresAt: entry.ExpiresAt}, nil
}

// Set stores a value under key with the TTL for cacheType.
func (c *Cache) Set(key, cacheType, valueJSON string) error {
	return c.store.SetCacheEntry(&db.CacheEntry{
		Key:       key,
		Type:      cacheType,
		ValueJSON: valueJSON,
		ExpiresAt: time.Now().Add(c.TTLFor(cacheType)),
	})
}

// PurgeExpired deletes expired entries and returns the count removed.
func (c *Cache) PurgeExpired() (int64, error) {
	return c.store.DeleteExpiredCache(time.Now())
}

// Stats combines persisted counts with in-process hit/miss counters.
type Stats struct {
	TotalKeywordsCached int64      `json:"totalKeywordsCached"`
	CacheHits           int64      `json:"cacheHits"`
	CacheMisses         int64      `json:"cacheMisses"`
	OldestMetricDate    *time.Time `json:"oldestMetricDate"`
	NewestMetricDate    *time.Time `json:"newestMetricDate"`
	ExpiredCacheEntries int64      `json:"expiredCacheEntries"`
	TotalCacheEntries   int64      `json:"totalCacheEntries"`
}

// Stats returns current cache statistics.
func (c *Cache) Stats() (Stats, error) {
	dbStats, err := c.store.CacheDBStats(time.Now())
	if err != nil {
		return Stats{}, err
	}
	return Stats{
		TotalKeywordsCached: dbStats.TotalKeywords,
		CacheHits:           c.hits.Load(),
		CacheMisses:         c.misses.Load(),
		OldestMetricDate:    dbStats.OldestMetricDate,
		NewestMetricDate:    dbStats.NewestMetricDate,
		ExpiredCacheEntries: dbStats.ExpiredCacheEntries,
		TotalCacheEntries:   dbStats.TotalCacheEntries,
	}, nil
}
