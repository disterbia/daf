package model

import (
	"gorm.io/gorm"
)

type ClinicalFeature struct {
	gorm.Model       // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Name             string
	Code             string
	JointActions     []JointAction     `gorm:"foreignKey:BodyCompositionId"`
	Recommendeds     []Recommended     `gorm:"foreignKey:ClinicalFeatureId"`
	UserJointActions []UserJointAction `gorm:"foreignKey:ClinicalFeatureId"`
}
