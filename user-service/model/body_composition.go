package model

import (
	"gorm.io/gorm"
)

type BodyComposition struct {
	gorm.Model   // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Name         string
	JointActions []JointAction `gorm:"foreignKey:BodyCompositionId"`
}
