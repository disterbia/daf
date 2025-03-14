package model

import (
	"gorm.io/gorm"
)

type CartItem struct {
	gorm.Model
	CartID          uint          `gorm:"index"`
	Cart            Cart          `gorm:"foreignKey:CartID"`
	ProductID       uint          `gorm:"index"`
	Product         Product       `gorm:"foreignKey:ProductID"`
	ProductOptionID uint          `gorm:"index"`
	ProductOption   ProductOption `gorm:"foreignKey:ProductOptionID"`
	Quantity        uint
}
