package model

import "gorm.io/gorm"

type Recommended struct {
	gorm.Model
	Exercise          Exercise `gorm:"foreignKey:ExerciseId"`
	ExerciseId        uint
	BodyComposition   BodyComposition `gorm:"foreignKey:BodyCompositionId"`
	BodyCompositionId uint
	Rom               Rom `gorm:"foreignKey:RomId"`
	RomId             uint
	ClinicalFeature   ClinicalFeature `gorm:"foreignKey:ClinicalFeatureId"`
	ClinicalFeatureId uint
	Degree            Degree `gorm:"foreignKey:DegreeId"`
	DegreeId          uint
}
