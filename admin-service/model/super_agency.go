package model

import (
	"gorm.io/gorm"
)

type SuperAgency struct {
	gorm.Model // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Name       string
	Agencies   []Agency `gorm:"foreignKey:SuperAgencyID"`
}
