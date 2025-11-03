package model

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID    int64          `gorm:"index;not null" json:"user_id"`
	Title     string         `gorm:"type:varchar(100)" json:"title"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
