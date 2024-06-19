package model

import (
	"gorm.io/gorm"
)

type UserDisableDetail struct {
	gorm.Model
	User            User          `gorm:"foreignKey:UID"`
	UID             uint          `gorm:"index"`
	DisableDetail   DisableDetail `gorm:"foreignKey:DisableDetailID"`
	DisableDetailID uint          `gorm:"index"`
}
