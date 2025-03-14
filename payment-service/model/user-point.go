package model

import (
	"gorm.io/gorm"
)

type UserPoint struct {
	gorm.Model
	User  User `gorm:"foreignKey:Uid"`
	Uid   uint
	Point uint
	Memo  string
}
