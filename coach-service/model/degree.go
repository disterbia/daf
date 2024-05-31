package model

import (
	"gorm.io/gorm"
)

type Degree struct {
	gorm.Model       // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Min              uint
	Max              uint
	Recommendeds     []Recommended     `gorm:"foreignKey:DegreeID"`
	UserJointActions []UserJointAction `gorm:"foreignKey:DegreeID"`
	Historys         []History         `gorm:"foreignKey:DegreeID"`
}
