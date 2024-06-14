package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model    // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Name          string
	Birthday      time.Time `gorm:"type:date"`
	Phone         string    `gorm:"unique"`
	Gender        bool      // true:남 false: 여
	Addr          string    `json:"addr"`
	AddrDetail    string    `json:"addr_detail"`
	Memo          string    `json:"memo"`
	Agency        Agency    `gorm:"foreignKey:AgencyID"`
	AgencyID      uint
	Admin         Admin `gorm:"foreignKey:AdminID"`
	AdminID       uint
	RegistDay     time.Time `gorm:"type:date"`
	UseStatus     UseStatus `gorm:"foreignKey:UseStatusID"`
	UseStatusID   uint
	CreateAdmin   Admin `gorm:"foreignKey:CreateAdminID"`
	CreateAdminID uint

	Email            *string `gorm:"unique"`
	Nickname         string
	DeviceID         string
	FCMToken         string
	SnsType          uint
	Images           []Image           `gorm:"foreignKey:Uid"`
	UserJointActions []UserJointAction `gorm:"foreignKey:Uid"`
}
