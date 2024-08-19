package model

import (
	"gorm.io/gorm"
)

type Admin struct {
	gorm.Model
	Email       string `gorm:"unique"`
	Password    string
	Agency      Agency `gorm:"foreignKey:AgencyID"`
	AgencyID    uint   `gorm:"index"`
	Name        string
	EnglishName string
	Phone       string `gorm:"unique"`
	Tel         string
	Fax         string
	IsApproval  bool
	Role        Role `gorm:"foreignKey:RoleID"`
	RoleID      uint `gorm:"index"`
}
