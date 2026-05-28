package db

import (
	"gorm.io/gorm"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// AutoMigrate creates/updates tables for all models. This is sufficient for the
// MVP; the single entry point keeps a clean seam for introducing versioned
// migrations later.
func AutoMigrate(gdb *gorm.DB) error {
	if err := gdb.AutoMigrate(AllModels()...); err != nil {
		return utils.NewDatabaseError(err)
	}
	return nil
}
