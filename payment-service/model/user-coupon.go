package model

import (
	"gorm.io/gorm"
)

type UserCoupon struct {
	gorm.Model
	User     User `gorm:"foreignKey:Uid"`
	Uid      uint
	Coupon   Coupon `gorm:"foreignKey:CouponId"`
	CouponId uint
}
