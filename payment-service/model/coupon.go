package model

import (
	"time"

	"gorm.io/gorm"
)

type Coupon struct {
	gorm.Model
	Name               string
	Detail             string
	Price              uint // percent가 0일떄 적용
	Percent            uint // price가 0 일때 적용
	DueDate            time.Time
	MinimumPrice       uint //사용 가능한 최소 금액
	MaximumPrice       uint //percent로 할인시 최대 금액
	CanDouble          bool //중복적용 가능 여부
	PossibleProductIds []uint
}
