package model

import (
	"gorm.io/gorm"
)

type Cart struct {
	gorm.Model
	User            User          `gorm:"foreignKey:Uid"`
	Uid             uint          `gorm:"index"`
	ProductID       uint          `gorm:"index"`
	Product         Product       `gorm:"foreignKey:ProductID"`
	ProductOptionID uint          `gorm:"index"`
	ProductOption   ProductOption `gorm:"foreignKey:ProductOptionID"`
	Quantity        uint
}
