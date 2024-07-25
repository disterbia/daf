package model

import "gorm.io/gorm"

type Image struct {
	gorm.Model
	User User `gorm:"foreignKey:Uid"`
	Uid  uint

	//부모 아이디
	ParentId uint
	Type     uint

	Url          string
	ThumbnailUrl string
}
