package model

import "gorm.io/gorm"

type AuthCode struct {
	gorm.Model
	Email string
	Code  string
}
