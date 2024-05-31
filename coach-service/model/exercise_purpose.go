package model

import (
	"gorm.io/gorm"
)

type ExercisePurpose struct {
	gorm.Model
	Exercise   Exercise `gorm:"foreignKey:ExerciseID"`
	ExerciseID uint
	Purpose    Purpose `gorm:"foreignKey:PurposeID"`
	PurposeID  uint
}
