package model

import (
	"gorm.io/gorm"
)

type UserVisit struct {
	gorm.Model
	User           User         `gorm:"foreignKey:UID"`
	UID            uint         `gorm:"index"`
	VisitPurpose   VisitPurpose `gorm:"foreignKey:VisitPurposeID"`
	VisitPurposeID uint         `gorm:"index"`
}
