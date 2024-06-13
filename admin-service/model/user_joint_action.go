package model

import (
	"gorm.io/gorm"
)

type UserJointAction struct {
	gorm.Model
	Name          string
	User          User `gorm:"foreignKey:Uid"`
	Uid           uint
	JointActions  JointAction `gorm:"foreignKey:JointActionId"`
	JointActionId JointAction

	RomId             uint
	ClinicalFeature   ClinicalFeature `gorm:"foreignKey:ClinicalFeatureId"`
	ClinicalFeatureId uint

	DegreeId uint
}
