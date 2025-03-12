package model

import (
	"gorm.io/gorm"
)

type UserRecommend struct {
	gorm.Model
	ToUser     User `gorm:"foreignKey:ToUserID"` // 추천을 받는 유저
	ToUserID   uint `gorm:"index"`
	FromUser   User `gorm:"foreignKey:FromUserID"` // 추천을 한 유저
	FromUserID uint `gorm:"index"`
}
