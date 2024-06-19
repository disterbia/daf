package model

import (
	"gorm.io/gorm"
)

type UserAfcHistoryGroup struct {
	gorm.Model
	Admin   Admin `gorm:"foreignKey:AdminID"`
	AdminID uint
}
