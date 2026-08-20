package database

import (
	"database/sql"
	"time"

	"github.com/acme/certpilot/internal/shared/config"
)

func configurePool(db *sql.DB, cfg config.DatabaseConfig) {
	if cfg.Driver == "sqlite" {
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
		return
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(config.ParseDurationDefault(cfg.ConnMaxLifetime, 30*time.Minute))
}
