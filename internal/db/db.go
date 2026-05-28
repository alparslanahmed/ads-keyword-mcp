package db

import (
	"log/slog"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// Open opens (and creates if needed) the SQLite database and runs migrations.
// The pure-Go SQLite driver is used so no CGO toolchain is required.
func Open(dsn string, logger *slog.Logger) (*gorm.DB, error) {
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, utils.NewDatabaseError(err)
	}
	if err := AutoMigrate(gdb); err != nil {
		return nil, err
	}
	if logger != nil {
		logger.Debug("database opened", "dsn", dsn)
	}
	return gdb, nil
}

// Close releases the underlying connection pool.
func Close(gdb *gorm.DB) error {
	sqlDB, err := gdb.DB()
	if err != nil {
		return utils.NewDatabaseError(err)
	}
	return sqlDB.Close()
}
