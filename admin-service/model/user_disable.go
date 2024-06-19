package model

import (
	"gorm.io/gorm"
)

type UserDisable struct {
	gorm.Model
	User          User        `gorm:"foreignKey:UID"`
	UID           uint        `gorm:"index"`
	DisableType   DisableType `gorm:"foreignKey:DisableTypeID"`
	DisableTypeID uint        `gorm:"index"`
}
