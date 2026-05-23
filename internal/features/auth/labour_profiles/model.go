package labour_profiles

import (
	"time"

	"gorm.io/gorm"
)

type LabourProfile struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"uniqueIndex;not null"`
	Trade      string         `json:"trade" gorm:"size:100"`
	Experience int            `json:"experience" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
