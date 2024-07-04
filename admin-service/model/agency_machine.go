package model

import (
	"gorm.io/gorm"
)

type AgencyMachine struct {
	gorm.Model
	Agency    Agency  `gorm:"foreignKey:AgencyID"`
	AgencyID  uint    `gorm:"index"`
	Machine   Machine `gorm:"foreignKey:MachineID"`
	MachineID uint    `gorm:"index"`
}
