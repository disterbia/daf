package model

import (
	"gorm.io/gorm"
)

type UserAfcHistory struct {
	gorm.Model
	Name                  string
	BodyComposition       BodyComposition     `gorm:"foreignKey:BodyCompositionID"`
	BodyCompositionID     uint                `gorm:"index"`
	JointAction           JointAction         `gorm:"foreignKey:JointActionID"`
	JointActionID         *uint               `gorm:"index"`
	Rom                   Rom                 `gorm:"foreignKey:RomID"`
	RomID                 *uint               `gorm:"index"`
	ClinicalFeature       ClinicalFeature     `gorm:"foreignKey:ClinicalFeatureID"`
	ClinicalFeatureID     *uint               `gorm:"index"`
	Degree                Degree              `gorm:"foreignKey:DegreeID"`
	DegreeID              *uint               `gorm:"index"`
	UserAfcHistoryGroup   UserAfcHistoryGroup `gorm:"foreignKey:UserAfcHistoryGroupID"`
	UserAfcHistoryGroupID uint                `gorm:"index"`
	Admin                 Admin               `gorm:"foreignKey:AdminID"`
	AdminID               uint                `gorm:"index"`
	IsGrip                *bool
	Pain                  uint
}
