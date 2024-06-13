package model

import (
	"gorm.io/gorm"
)

type Exercise struct {
	gorm.Model
	Name       string `gorm:"unique"`
	CategoryId uint
}
