package model

import "gorm.io/gorm"

type VerifiedEmail struct {
	gorm.Model
	Email string
}
