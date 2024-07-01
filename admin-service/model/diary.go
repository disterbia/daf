package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Diary struct {
	gorm.Model // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Title      string
	User       User      `gorm:"foreignKey:Uid"`
	Uid        uint      `gorm:"index"`
	ClassDate  time.Time `gorm:"type:date"`
	ClassName  string
	ClassType  uint
	Explain    json.RawMessage `gorm:"type:jsonb"`
	Admin      Admin           `gorm:"foreignKey:AdminID"`
	AdminID    uint            `gorm:"index"`
}
