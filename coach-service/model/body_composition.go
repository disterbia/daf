package model

import (
	"gorm.io/gorm"
)

type BodyComposition struct {
	gorm.Model   // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Name         string
	BodyType     BodyType `gorm:"foreignKey:BodyTypeID"`
	BodyTypeID   uint
	JointActions []JointAction `gorm:"foreignKey:BodyCompositionID"`
}
