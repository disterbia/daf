package model

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Exercise struct {
	gorm.Model
	Name       string          `gorm:"unique"`
	Category   Category        `gorm:"foreignKey:CategoryId"`
	CategoryId uint            `gorm:"index"`
	Explain    json.RawMessage `gorm:"type:jsonb"`
}
