package model

import (
	"gorm.io/gorm"
)

type UserJointAction struct {
	gorm.Model
	Name string
	User User `gorm:"foreignKey:Uid"`
	Uid  uint `gorm:"index"`

	JointActionID uint `gorm:"index"`
	Rom           Rom  `gorm:"foreignKey:RomID"`
	RomID         uint `gorm:"index"`

	ClinicalFeatureID uint `gorm:"index"`

	DegreeID uint `gorm:"index"`
}
