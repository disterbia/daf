package model

import "gorm.io/gorm"

type Image struct {
	gorm.Model
	Uid  uint
	User User `gorm:"foreignKey:Uid"`
	//부모 아이디
	ParentId uint `json:"parent_id"`
	Type     uint

	Url          string
	ThumbnailUrl string `json:"thumbnail_url"`
}
