package model

import (
	"gorm.io/gorm"
)

type RecommendedJointRom struct {
	gorm.Model                // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Recommended   Recommended `gorm:"foreignKey:RecommendedID"`
	RecommendedID uint        `gorm:"index"`
	JointAction   JointAction `gorm:"foreignKey:JointActionID"`
	JointActionID uint        `gorm:"index"`
	Rom           Rom         `gorm:"foreignKey:RomID"`
	RomID         uint        `gorm:"index"`
}
