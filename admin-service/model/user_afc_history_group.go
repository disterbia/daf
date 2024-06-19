package model

import (
	"gorm.io/gorm"
)

type UserAfcHistoryGroup struct {
	gorm.Model
	User    User `gorm:"foreignKey:Uid"`
	Uid     uint
	Admin   Admin `gorm:"foreignKey:AdminID"`
	AdminID uint
}
