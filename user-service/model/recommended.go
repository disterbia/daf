package model

import "gorm.io/gorm"

type Recommended struct {
	gorm.Model
	Exercise          Exercise `gorm:"foreignKey:ExerciseID"`
	ExerciseID        uint     `gorm:"index"`
	BodyFilter        uint
	BodyType          BodyType        `gorm:"foreignKey:BodyTypeID"`
	BodyTypeID        uint            `gorm:"index"`
	Rom               Rom             `gorm:"foreignKey:RomID"`
	RomID             uint            `gorm:"index"`
	ClinicalFeature   ClinicalFeature `gorm:"foreignKey:ClinicalFeatureID"`
	ClinicalFeatureID uint            `gorm:"index"`
	Degree            Degree          `gorm:"foreignKey:DegreeId"`
	DegreeId          uint            `gorm:"index"`
	IsAsymmetric      bool
}
