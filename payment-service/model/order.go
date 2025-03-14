package model

import (
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	User         User `gorm:"foreignKey:Uid"`
	Uid          uint `gorm:"index"`
	OriginPrice  uint
	OrderPrice   uint
	Point        uint
	OrderItems   []OrderItem   `gorm:"foreignKey:OrderId"`
	OrderCoupons []OrderCoupon `gorm:"foreignKey:OrderId"`
	Tid          string        `gorm:"unique"`
}
