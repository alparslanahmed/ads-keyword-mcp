package db

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// Store is the data-access layer over GORM. All persistence flows through here.
type Store struct {
	db *gorm.DB
}

// NewStore wraps a *gorm.DB.
func NewStore(gdb *gorm.DB) *Store { return &Store{db: gdb} }

// DB exposes the underlying handle for advanced callers (kept minimal).
func (s *Store) DB() *gorm.DB { return s.db }

// ---- ResearchRun ----

func (s *Store) CreateRun(run *ResearchRun) error {
	if run.ID == "" {
		run.ID = uuid.NewString()
	}
	if err := s.db.Create(run).Error; err != nil {
		return utils.NewDatabaseError(err)
	}
	return nil
}

func (s *Store) SaveRun(run *ResearchRun) error {
	if err := s.db.Save(run).Error; err != nil {
		return utils.NewDatabaseError(err)
	}
	return nil
}

func (s *Store) GetRun(id string) (*ResearchRun, error) {
	var run ResearchRun
	err := s.db.First(&run, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, utils.NewDatabaseError(err)
	}
	return &run, nil
}

func (s *Store) ListRecentRuns(limit int) ([]ResearchRun, error) {
	if limit <= 0 {
		limit = 20
	}
	var runs []ResearchRun
	if err := s.db.Order("created_at DESC").Limit(limit).Find(&runs).Error; err != nil {
		return nil, utils.NewDatabaseError(err)
	}
	return runs, nil
}

// ---- Keyword ----

// UpsertKeyword finds an existing keyword by normalized text or creates one.
func (s *Store) UpsertKeyword(text, normalized string) (*Keyword, error) {
	var kw Keyword
	err := s.db.First(&kw, "normalized_text = ?", normalized).Error
	if err == nil {
		return &kw, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.NewDatabaseError(err)
	}
	kw = Keyword{ID: uuid.NewString(), Text: text, NormalizedText: normalized, CreatedAt: time.Now()}
	// OnConflict guards against a race where another writer inserts the same
	// normalized text concurrently.
	res := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "normalized_text"}},
		DoNothing: true,
	}).Create(&kw)
	if res.Error != nil {
		return nil, utils.NewDatabaseError(res.Error)
	}
	if res.RowsAffected == 0 {
		if err := s.db.First(&kw, "normalized_text = ?", normalized).Error; err != nil {
			return nil, utils.NewDatabaseError(err)
		}
	}
	return &kw, nil
}

// GetKeywordsByIDs returns keywords keyed by ID for the given IDs.
func (s *Store) GetKeywordsByIDs(ids []string) (map[string]Keyword, error) {
	out := make(map[string]Keyword)
	if len(ids) == 0 {
		return out, nil
	}
	var kws []Keyword
	if err := s.db.Where("id IN ?", ids).Find(&kws).Error; err != nil {
		return nil, utils.NewDatabaseError(err)
	}
	for _, k := range kws {
		out[k.ID] = k
	}
	return out, nil
}

// ---- KeywordMetric ----

func (s *Store) CreateMetric(m *KeywordMetric) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if err := s.db.Create(m).Error; err != nil {
		return utils.NewDatabaseError(err)
	}
	return nil
}

// GetMetricsByIDs returns metrics keyed by ID for the given IDs.
func (s *Store) GetMetricsByIDs(ids []string) (map[string]KeywordMetric, error) {
	out := make(map[string]KeywordMetric)
	if len(ids) == 0 {
		return out, nil
	}
	var metrics []KeywordMetric
	if err := s.db.Where("id IN ?", ids).Find(&metrics).Error; err != nil {
		return nil, utils.NewDatabaseError(err)
	}
	for _, m := range metrics {
		out[m.ID] = m
	}
	return out, nil
}

// ---- ResearchRunKeyword ----

func (s *Store) CreateRunKeyword(rk *ResearchRunKeyword) error {
	if rk.ID == "" {
		rk.ID = uuid.NewString()
	}
	if err := s.db.Create(rk).Error; err != nil {
		return utils.NewDatabaseError(err)
	}
	return nil
}

func (s *Store) ListRunKeywords(runID string) ([]ResearchRunKeyword, error) {
	var rks []ResearchRunKeyword
	if err := s.db.Where("research_run_id = ?", runID).
		Order("opportunity_score DESC").Find(&rks).Error; err != nil {
		return nil, utils.NewDatabaseError(err)
	}
	return rks, nil
}

// ---- ApiCallLog ----

func (s *Store) LogApiCall(l *ApiCallLog) error {
	if l.ID == "" {
		l.ID = uuid.NewString()
	}
	if err := s.db.Create(l).Error; err != nil {
		return utils.NewDatabaseError(err)
	}
	return nil
}

// ---- CacheEntry ----

func (s *Store) GetCacheEntry(key string) (*CacheEntry, error) {
	var e CacheEntry
	err := s.db.First(&e, "key = ?", key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, utils.NewCacheError(err)
	}
	return &e, nil
}

func (s *Store) SetCacheEntry(e *CacheEntry) error {
	now := time.Now()
	e.UpdatedAt = now
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	err := s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"type", "value_json", "expires_at", "updated_at",
		}),
	}).Create(e).Error
	if err != nil {
		return utils.NewCacheError(err)
	}
	return nil
}

func (s *Store) DeleteExpiredCache(now time.Time) (int64, error) {
	res := s.db.Where("expires_at < ?", now).Delete(&CacheEntry{})
	if res.Error != nil {
		return 0, utils.NewCacheError(res.Error)
	}
	return res.RowsAffected, nil
}

// ---- Stats ----

// CacheDBStats are the database-derived portion of cache statistics.
type CacheDBStats struct {
	TotalKeywords       int64
	TotalCacheEntries   int64
	ExpiredCacheEntries int64
	OldestMetricDate    *time.Time
	NewestMetricDate    *time.Time
}

func (s *Store) CacheDBStats(now time.Time) (CacheDBStats, error) {
	var out CacheDBStats
	if err := s.db.Model(&Keyword{}).Count(&out.TotalKeywords).Error; err != nil {
		return out, utils.NewDatabaseError(err)
	}
	if err := s.db.Model(&CacheEntry{}).Count(&out.TotalCacheEntries).Error; err != nil {
		return out, utils.NewDatabaseError(err)
	}
	if err := s.db.Model(&CacheEntry{}).Where("expires_at < ?", now).
		Count(&out.ExpiredCacheEntries).Error; err != nil {
		return out, utils.NewDatabaseError(err)
	}
	var oldest, newest KeywordMetric
	if err := s.db.Order("fetched_at ASC").First(&oldest).Error; err == nil {
		t := oldest.FetchedAt
		out.OldestMetricDate = &t
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return out, utils.NewDatabaseError(err)
	}
	if err := s.db.Order("fetched_at DESC").First(&newest).Error; err == nil {
		t := newest.FetchedAt
		out.NewestMetricDate = &t
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return out, utils.NewDatabaseError(err)
	}
	return out, nil
}
