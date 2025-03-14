package model

import (
	"gorm.io/gorm"
)

type CouponHistory struct {
	gorm.Model
	User         User `gorm:"foreignKey:Uid"`
	Uid          uint
	UserCoupon   UserCoupon `gorm:"foreignKey:UserCouponId"`
	UserCouponId uint
}
