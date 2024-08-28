package model

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Exercise struct {
	gorm.Model
	Name    string          `gorm:"unique"`
	Explain json.RawMessage `gorm:"type:jsonb"`
}
