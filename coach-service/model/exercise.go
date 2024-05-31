package model

import (
	"gorm.io/gorm"
)

type Exercise struct {
	gorm.Model
	Name             string   `gorm:"unique"`
	Category         Category `gorm:"foreignKey:CategoryID"`
	CategoryID       uint
	ExerciseMachines []ExerciseMachine `gorm:"foreignKey:ExerciseID"`
	Recommendeds     []Recommended     `gorm:"foreignKey:ExerciseID"`
	Historys         []History         `gorm:"foreignKey:ExerciseID"`
	ExercisePurpose  []ExercisePurpose `gorm:"foreignKey:ExerciseID"`
}
