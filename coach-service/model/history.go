package model

import "gorm.io/gorm"

type History struct {
	gorm.Model
	Exercise   Exercise `gorm:"foreignKey:ExerciseID"`
	ExerciseID uint     `gorm:"index"`

	JointActionID uint `gorm:"index"`
	Rom           Rom  `gorm:"foreignKey:RomID"`
	RomID         uint `gorm:"index"`

	ClinicalFeatureID uint `gorm:"index"`

	DegreeID uint `gorm:"index"`
}
