package model

import (
	"gorm.io/gorm"
)

type Admin struct {
	gorm.Model
	Email       string `gorm:"unique"`
	Password    string
	Agency      Agency `gorm:"foreignKey:AgencyID"`
	AgencyID    uint
	Name        string
	EnglishName string
	Phone       string `gorm:"unique"`
	Tel         string `gorm:"unique"`
	Fax         string `gorm:"unique"`
	IsApproval  bool
	Role        Role `gorm:"foreignKey:RoleID"`
	RoleID      uint
}
