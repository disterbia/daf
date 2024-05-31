package model

import (
	"gorm.io/gorm"
)

type BodyType struct {
	gorm.Model
	Name             string
	BodyCompositions []BodyComposition `gorm:"foreignKey:BodyTypeID"`
	Recommendeds     []Recommended     `gorm:"foreignKey:BodyTypeID"`
}
