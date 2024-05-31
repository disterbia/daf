package model

import "gorm.io/gorm"

type History struct {
	gorm.Model
	Exercise          Exercise `gorm:"foreignKey:ExerciseID"`
	ExerciseID        uint
	JointAction       JointAction `gorm:"foreignKey:JointActionID"`
	JointActionID     uint
	Rom               Rom `gorm:"foreignKey:RomID"`
	RomID             uint
	ClinicalFeature   ClinicalFeature `gorm:"foreignKey:ClinicalFeatureID"`
	ClinicalFeatureID uint
	Degree            Degree `gorm:"foreignKey:DegreeID"`
	DegreeID          uint
}
