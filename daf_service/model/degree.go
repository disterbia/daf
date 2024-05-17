package model

import (
	"gorm.io/gorm"
)

type Degree struct {
	gorm.Model       // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Min              uint
	Max              uint
	Recommendeds     []Recommended     `gorm:"foreignKey:DegreeId"`
	UserJointActions []UserJointAction `gorm:"foreignKey:DegreeId"`
}
