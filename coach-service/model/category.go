package model

import (
	"gorm.io/gorm"
)

type Category struct {
	gorm.Model
	Name      string
	Exercises []Exercise `json:"exercises" gorm:"foreignKey:CategoryID"`
}
