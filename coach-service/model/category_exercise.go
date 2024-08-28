package model

import (
	"gorm.io/gorm"
)

type CategoryExercise struct {
	gorm.Model
	Exercise   Exercise `gorm:"foreignKey:ExerciseID"`
	ExerciseID uint     `gorm:"index"`
	Category   Category `gorm:"foreignKey:CategoryID"`
	CategoryID uint     `gorm:"index"`
}
