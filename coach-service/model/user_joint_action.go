package model

import (
	"gorm.io/gorm"
)

type UserJointAction struct {
	gorm.Model
	Name              string
	User              User            `gorm:"foreignKey:Uid"`
	Uid               uint            `gorm:"index"`
	JointAction       JointAction     `gorm:"foreignKey:JointActionID"`
	JointActionID     uint            `gorm:"index"`
	Rom               Rom             `gorm:"foreignKey:RomID"`
	RomID             uint            `gorm:"index"`
	ClinicalFeature   ClinicalFeature `gorm:"foreignKey:ClinicalFeatureID"`
	ClinicalFeatureID uint            `gorm:"index"`
	Degree            Degree          `gorm:"foreignKey:DegreeID"`
	DegreeID          uint            `gorm:"index"`
}
