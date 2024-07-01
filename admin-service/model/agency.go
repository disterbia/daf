package model

import (
	"gorm.io/gorm"
)

type Agency struct {
	gorm.Model    // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Name          string
	Tel           string      `gorm:"unique"`
	SuperAgency   SuperAgency `gorm:"foreignKey:SuperAgencyID"`
	SuperAgencyID uint
	Admins        []Admin `gorm:"foreignKey:AgencyID"`
	Latitude      float64
	Longitude     float64
}
