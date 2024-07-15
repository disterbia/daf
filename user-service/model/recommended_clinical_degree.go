package model

import (
	"gorm.io/gorm"
)

type RecommendedClinicalDegree struct {
	gorm.Model                        // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Recommended       Recommended     `gorm:"foreignKey:RecommendedID"`
	RecommendedID     uint            `gorm:"index"`
	JointAction       JointAction     `gorm:"foreignKey:JointActionID"`
	JointActionID     uint            `gorm:"index"`
	ClinicalFeature   ClinicalFeature `gorm:"foreignKey:ClinicalFeatureID"`
	ClinicalFeatureID uint            `gorm:"index"`
	Degree            Degree          `gorm:"foreignKey:DegreeID"`
	DegreeID          uint            `gorm:"index"`
}
