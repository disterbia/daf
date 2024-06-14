package model

import (
	"gorm.io/gorm"
)

type UserJointAction struct {
	gorm.Model
	Name              string
	User              User `gorm:"foreignKey:Uid"`
	Uid               uint
	JointActions      JointAction `gorm:"foreignKey:JointActionId"`
	JointActionId     uint
	Rom               Rom `gorm:"foreignKey:RomId"`
	RomId             uint
	ClinicalFeature   ClinicalFeature `gorm:"foreignKey:ClinicalFeatureId"`
	ClinicalFeatureId uint
	Degree            Degree `gorm:"foreignKey:DegreeId"`
	DegreeId          uint
}
