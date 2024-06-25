package model

import (
	"gorm.io/gorm"
)

type ExerciseMeasure struct {
	gorm.Model
	Exercise   Exercise `gorm:"foreignKey:ExerciseID"`
	ExerciseID uint     `gorm:"index"`
	Measure    Measure  `gorm:"foreignKey:MeasureID"`
	MeasureID  uint     `gorm:"index"`
}
