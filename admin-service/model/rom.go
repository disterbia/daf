package model

import (
	"gorm.io/gorm"
)

type Rom struct {
	gorm.Model // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Min        *uint
	Max        *uint
	Name       *string
}
