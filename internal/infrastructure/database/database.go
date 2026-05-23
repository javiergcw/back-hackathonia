package database

import (
	"fmt"

	"github.com/javierg/hackathon-bqia/internal/features/auth/builder_profiles"
	"github.com/javierg/hackathon-bqia/internal/features/auth/labour_profiles"
	"github.com/javierg/hackathon-bqia/internal/features/auth/user"
	"github.com/javierg/hackathon-bqia/internal/infrastructure/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&user.User{},
		&builder_profiles.BuilderProfile{},
		&labour_profiles.LabourProfile{},
	)
}
