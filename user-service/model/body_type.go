package model

import (
	"gorm.io/gorm"
)

type BodyType struct {
	gorm.Model
	Name string
}
