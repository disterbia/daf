package model

import (
	"gorm.io/gorm"
)

type UserVisit struct {
	gorm.Model
	User           User `gorm:"foreignKey:UID"`
	UID            uint
	VisitPurpose   VisitPurpose `gorm:"foreignKey:VisitPurposeID"`
	VisitPurposeID uint
}
