package model

import (
	"gorm.io/gorm"
)

type Admin struct {
	gorm.Model
	Email    string `gorm:"unique"`
	Password string
	AgencyID uint `json:"agency_id"`
}
