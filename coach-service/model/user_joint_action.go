package model

import (
	"gorm.io/gorm"
)

type UserJointAction struct {
	gorm.Model
	Name              string
	User              User `gorm:"foreignKey:Uid"`
	Uid               uint
	JointAction       JointAction `gorm:"foreignKey:JointActionID"`
	JointActionID     uint
	Rom               Rom `gorm:"foreignKey:RomID"`
	RomID             uint
	ClinicalFeature   ClinicalFeature `gorm:"foreignKey:ClinicalFeatureID"`
	ClinicalFeatureID uint
	Degree            Degree `gorm:"foreignKey:DegreeID"`
	DegreeID          uint
}
