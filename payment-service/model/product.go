package model

import (
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name      string
	Service   Service `gorm:"foreignKey:ServiceId"`
	ServiceId uint    `gorm:"index"`
	Price     uint
	SellPrice uint
	Detail    string
	Options   []ProductOption `gorm:"foreignKey:ProductId"`
}
