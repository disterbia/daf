package model

import (
	"gorm.io/gorm"
)

type Recommended struct {
	gorm.Model
	Exercise        Exercise `gorm:"foreignKey:ExerciseID"`
	ExerciseID      uint     `gorm:"index"`
	BodyType        BodyType `gorm:"foreignKey:BodyTypeID"`
	BodyTypeID      uint     `gorm:"index"`
	LocoRom         Rom      `gorm:"foreignKey:LocoRomID"`
	LocoRomID       uint     `gorm:"index"`
	IsAsymmetric    bool
	IsGrip          *bool
	ClinicalDegrees []RecommendedClinicalDegree `gorm:"foreignKey:RecommendedID;constraint:OnDelete:CASCADE;"`
	JointRoms       []RecommendedJointRom       `gorm:"foreignKey:RecommendedID;constraint:OnDelete:CASCADE;"`
}
