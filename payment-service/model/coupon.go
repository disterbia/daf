package model

import (
	"time"

	"gorm.io/gorm"
)

type Coupon struct {
	gorm.Model
	Name         string
	Price        uint // MaximumPrice 대용으로 사용가능
	Percent      uint
	DueDate      time.Time
	MinimumPrice uint
}
