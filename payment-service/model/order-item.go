package model

import (
	"gorm.io/gorm"
)

type OrderItem struct {
	gorm.Model
	Order     Order `gorm:"foreignKey:OrderId"`
	OrderId   uint  `gorm:"index"`
	Name      string
	Product   Product `gorm:"foreignKey:ProductId"`
	ProductId uint
	Price     uint
	BuyPrice  uint
}
