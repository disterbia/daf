package model

import (
	"gorm.io/gorm"
)

type ClassPurpose struct {
	gorm.Model
	Name string
}
