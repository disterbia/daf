package model

import "gorm.io/gorm"

type Image struct {
	gorm.Model
	Uid uint

	//부모 아이디
	ParentId uint
	Type     uint

	Url          string
	ThumbnailUrl string
}
