package model

import (
	"gorm.io/gorm"
)

type UserJointAction struct {
	gorm.Model
	Name          string
	User          User `gorm:"foreignKey:Uid"`
	Uid           uint
	JointAction   JointAction `gorm:"foreignKey:JointActionId"`
	JointActionId uint

	RomId             uint
	ClinicalFeature   ClinicalFeature `gorm:"foreignKey:ClinicalFeatureId"`
	ClinicalFeatureId uint

	DegreeId uint
}
