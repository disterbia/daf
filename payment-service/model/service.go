package model

import (
	"gorm.io/gorm"
)

type Service struct {
	gorm.Model
	Name string
}
