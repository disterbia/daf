package model

import (
	"gorm.io/gorm"
)

type ProductOption struct {
	gorm.Model
	Name      string
	Product   Product `gorm:"foreignKey:ProductId"`
	ProductId uint    `gorm:"index"`
	Price     uint
}
