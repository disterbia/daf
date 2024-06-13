package model

import "gorm.io/gorm"

type History struct {
	gorm.Model
	Exercise   Exercise `gorm:"foreignKey:ExerciseId"`
	ExerciseId uint

	JointActionId uint

	RomId uint

	ClinicalFeatureId uint

	DegreeId uint
}
