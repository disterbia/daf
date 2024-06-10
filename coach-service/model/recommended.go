package model

import "gorm.io/gorm"

type Recommended struct {
	gorm.Model
	Exercise          Exercise `gorm:"foreignKey:ExerciseID"`
	ExerciseID        uint
	BodyFilter        uint
	BodyType          BodyType `gorm:"foreignKey:BodyTypeID"`
	BodyTypeID        uint
	Rom               Rom `gorm:"foreignKey:RomID"`
	RomID             uint
	ClinicalFeature   ClinicalFeature `gorm:"foreignKey:ClinicalFeatureID"`
	ClinicalFeatureID *uint
	Degree            Degree `gorm:"foreignKey:DegreeID"`
	DegreeID          *uint
	Asymmetric        bool
}
