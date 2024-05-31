package model

import (
	"gorm.io/gorm"
)

type Rom struct {
	gorm.Model       // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Min              uint
	Man              uint
	Recommendeds     []Recommended     `gorm:"foreignKey:RomID"`
	UserJointActions []UserJointAction `gorm:"foreignKey:RomID"`
	Historys         []History         `gorm:"foreignKey:RomID"`
}
