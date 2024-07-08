package model

import (
	"gorm.io/gorm"
)

type UserAfcHistoryGroup struct {
	gorm.Model
	User    User  `gorm:"foreignKey:Uid"`
	Uid     uint  `gorm:"index"`
	Admin   Admin `gorm:"foreignKey:AdminID"`
	AdminID uint  `gorm:"index"`
}
