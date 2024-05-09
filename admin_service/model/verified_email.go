package model

import "gorm.io/gorm"

type VerifiedEmail struct {
	gorm.Model
	Id    uint
	Email string
}
