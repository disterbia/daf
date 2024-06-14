package model

import "gorm.io/gorm"

type History struct {
	gorm.Model
	Exercise          Exercise `gorm:"foreignKey:ExerciseId"`
	ExerciseId        uint
	JointActions      JointAction `gorm:"foreignKey:JointActionId"`
	JointActionId     uint
	Rom               Rom `gorm:"foreignKey:RomId"`
	RomId             uint
	ClinicalFeature   ClinicalFeature `gorm:"foreignKey:ClinicalFeatureId"`
	ClinicalFeatureId uint
	Degree            Degree `gorm:"foreignKey:DegreeId"`
	DegreeId          uint
}
