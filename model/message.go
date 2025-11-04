package model

import (
	"time"
)

type Message struct {
	ID        string `gorm:"primaryKey;type:varchar(36)"`
	SessionID string `gorm:"index;not null;type:varchar(36)"`
	UserName  string `gorm:"type:varchar(20)"`
	Content   string `gorm:"type:text"`
	IsUser    bool   `gorm:"not null;default:true"`
	CreatedAt time.Time
}

type History struct {
	IsUser  bool   `json:"is_user"`
	Content string `json:"content"`
}
