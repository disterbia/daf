package model

import (
	"gorm.io/gorm"
)

type Machine struct {
	gorm.Model
	Name        string `gorm:"unique"`
	MachineType uint   `json:"machine_type"`
	Memo        string
}
