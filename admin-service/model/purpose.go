package model

import (
	"gorm.io/gorm"
)

type Purpose struct {
	gorm.Model
	Name string `gorm:"unique"`
}
