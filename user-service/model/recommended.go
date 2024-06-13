package model

import "gorm.io/gorm"

type Recommended struct {
	gorm.Model
	Exercise   Exercise `gorm:"foreignKey:ExerciseId"`
	ExerciseId uint

	BodyCompositionId uint

	RomId uint

	ClinicalFeatureId uint

	DegreeId uint
}
