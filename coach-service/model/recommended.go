package model

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Recommended struct {
	gorm.Model
	Exercise          Exercise `gorm:"foreignKey:ExerciseID"`
	ExerciseID        uint     `gorm:"index"`
	BodyFilter        uint
	BodyType          BodyType        `gorm:"foreignKey:BodyTypeID"`
	BodyTypeID        uint            `gorm:"index"`
	Rom               Rom             `gorm:"foreignKey:RomID"`
	RomID             *uint           `gorm:"index"`
	ClinicalFeature   ClinicalFeature `gorm:"foreignKey:ClinicalFeatureID"`
	ClinicalFeatureID *uint           `gorm:"index"`
	Degree            Degree          `gorm:"foreignKey:DegreeID"`
	DegreeID          *uint           `gorm:"index"`
	IsAsymmetric      bool
	AmputationCode    uint
	Explain           json.RawMessage `gorm:"type:jsonb"`
}
