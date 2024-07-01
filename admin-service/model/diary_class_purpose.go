package model

import (
	"gorm.io/gorm"
)

type DiaryClassPurpose struct {
	gorm.Model
	ClassPurpose   ClassPurpose `gorm:"foreignKey:ClassPurposeID"`
	ClassPurposeID uint         `gorm:"index"`
	Diary          Diary        `gorm:"foreignKey:DiaryID"`
	DiaryID        uint         `gorm:"index"`
}
