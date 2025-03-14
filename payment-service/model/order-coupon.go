package model

import (
	"gorm.io/gorm"
)

type OrderCoupon struct {
	gorm.Model
	Order    Order  `gorm:"foreignKey:CouponId"`
	OrderId  uint   `gorm:"index"`
	Coupon   Coupon `gorm:"foreignKey:CouponId"`
	CouponId uint   `gorm:"index"`
}
