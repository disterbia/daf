package model

import "gorm.io/gorm"

type History struct {
	gorm.Model
	Exercise          Exercise        `gorm:"foreignKey:ExerciseId"`
	ExerciseId        uint            `gorm:"index"`
	BodyComposition   BodyComposition `gorm:"foreignKey:BodyCompositionID"`
	BodyCompositionID uint            `gorm:"index"`
	JointActions      JointAction     `gorm:"foreignKey:JointActionID"`
	JointActionID     uint            `gorm:"index"`
	Rom               Rom             `gorm:"foreignKey:RomID"`
	RomID             uint            `gorm:"index"`
	ClinicalFeature   ClinicalFeature `gorm:"foreignKey:ClinicalFeatureID"`
	ClinicalFeatureID uint            `gorm:"index"`
	Degree            Degree          `gorm:"foreignKey:DegreeId"`
	DegreeId          uint            `gorm:"index"`
}
