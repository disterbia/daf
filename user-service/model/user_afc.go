package model

import (
	"gorm.io/gorm"
)

type UserAfc struct {
	gorm.Model
	User              User `gorm:"foreignKey:Uid"`
	Uid               uint
	BodyComposition   BodyComposition `gorm:"foreignKey:BodyCompositionID"`
	BodyCompositionID uint
	JointAction       JointAction     `gorm:"foreignKey:JointActionID"`
	JointActionID     uint            `gorm:"index"`
	Rom               Rom             `gorm:"foreignKey:RomID"`
	RomID             uint            `gorm:"index"`
	ClinicalFeature   ClinicalFeature `gorm:"foreignKey:ClinicalFeatureID"`
	ClinicalFeatureID *uint           `gorm:"index"`
	Degree            Degree          `gorm:"foreignKey:DegreeId"`
	DegreeId          *uint           `gorm:"index"`
	Pain              uint
}
