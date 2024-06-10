package model

import (
	"gorm.io/gorm"
)

type ExerciseMachine struct {
	gorm.Model
	Exercise   Exercise `gorm:"foreignKey:ExerciseID"`
	ExerciseID uint     `gorm:"index"`
	Machine    Machine  `gorm:"foreignKey:MachineID"`
	MachineID  uint     `gorm:"index"`
}
