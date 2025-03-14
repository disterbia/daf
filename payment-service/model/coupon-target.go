package model

import (
	"gorm.io/gorm"
)

type CouponTarget struct {
	gorm.Model
	Coupon    Coupon  `gorm:"foreignKey:CouponId"`
	CouponId  uint    `gorm:"index"`
	Service   Service `gorm:"foreignKey:ServiceId"`
	ServiceId uint    `gorm:"index"`
}
