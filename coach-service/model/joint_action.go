package model

import (
	"gorm.io/gorm"
)

type JointAction struct {
	gorm.Model        // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Name              string
	BodyComposition   BodyComposition `gorm:"foreignKey:BodyCompositionID"`
	BodyCompositionID uint
	UserJointActions  []UserJointAction `gorm:"foreignKey:JointActionID"`
	Historys          []History         `gorm:"foreignKey:JointActionID"`
}
