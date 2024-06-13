package model

import "gorm.io/gorm"

type Recommended struct {
	gorm.Model
	Exercise   Exercise `gorm:"foreignKey:ExerciseID"`
	ExerciseID uint
	BodyFilter uint

	BodyTypeID uint

	RomID uint

	ClinicalFeatureID uint

	DegreeID   uint
	Asymmetric bool
}
