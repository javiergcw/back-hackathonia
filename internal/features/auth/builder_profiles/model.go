package builder_profiles

import (
	"time"

	"gorm.io/gorm"
)

type BuilderProfile struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"uniqueIndex;not null"`
	Company   string         `json:"company" gorm:"size:255"`
	License   string         `json:"license" gorm:"size:100"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
