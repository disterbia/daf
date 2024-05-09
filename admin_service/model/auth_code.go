package model

import "gorm.io/gorm"

type AuthCode struct {
	gorm.Model
	Id    uint
	Email string
	Code  string
}
