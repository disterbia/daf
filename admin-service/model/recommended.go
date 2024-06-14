package model

import "gorm.io/gorm"

type Recommended struct {
	gorm.Model
	Exercise          Exercise `gorm:"foreignKey:ExerciseID"`
	ExerciseID        uint
	BodyFilter        uint
	BodyType          BodyType `gorm:"foreignKey:BodyTypeID"`
	BodyTypeID        uint
	Rom               Rom `gorm:"foreignKey:RomId"`
	RomId             uint
	ClinicalFeature   ClinicalFeature `gorm:"foreignKey:ClinicalFeatureId"`
	ClinicalFeatureId uint
	Degree            Degree `gorm:"foreignKey:DegreeId"`
	DegreeId          uint
	Asymmetric        bool
}
