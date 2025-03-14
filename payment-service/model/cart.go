package model

import (
	"gorm.io/gorm"
)

type Cart struct {
	gorm.Model
	User            User          `gorm:"foreignKey:Uid"`
	Uid             uint          `gorm:"index"`
	ProductOptionID uint          `gorm:"index"`
	ProductOption   ProductOption `gorm:"foreignKey:ProductOptionID"`
	Quantity        uint
}
