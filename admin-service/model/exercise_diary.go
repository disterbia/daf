package model

import (
	"gorm.io/gorm"
)

type ExerciseDiary struct {
	gorm.Model
	Exercise   Exercise `gorm:"foreignKey:ExerciseID"`
	ExerciseID uint     `gorm:"index"`
	Measure    Measure  `gorm:"foreignKey:MeasureID"`
	MeasureID  uint     `gorm:"index"`
	Diary      Diary    `gorm:"foreignKey:DiaryID"`
	DiaryID    uint     `gorm:"index"`
	Value      uint
}
