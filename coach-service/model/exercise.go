package model

import (
	"gorm.io/gorm"
)

type Exercise struct {
	gorm.Model
	Name       string   `gorm:"unique"`
	Category   Category `gorm:"foreignKey:CategoryID"`
	CategoryID uint     `gorm:"index"`
}
