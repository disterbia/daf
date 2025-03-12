package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model            // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Email         *string `gorm:"unique"`
	SnsEmail      *string `gorm:"unique"`
	Username      *string `gorm:"unique"`
	SnsId         *string `gorm:"unique"`
	Password      *string
	Name          string
	Birthday      time.Time `gorm:"type:date"`
	Phone         string    `gorm:"unique"`
	Gender        uint
	Addr          string    `json:"addr"`
	AddrDetail    string    `json:"addr_detail"`
	Memo          string    `json:"memo"`
	Agency        Agency    `gorm:"foreignKey:AgencyID"`
	AgencyID      *uint     `gorm:"index"`
	Admin         Admin     `gorm:"foreignKey:AdminID"`
	AdminID       *uint     `gorm:"index"`
	RegistDay     time.Time `gorm:"type:date"`
	UseStatus     UseStatus `gorm:"foreignKey:UseStatusID"`
	UseStatusID   *uint     `gorm:"index"`
	CreateAdmin   Admin     `gorm:"foreignKey:CreateAdminID"`
	CreateAdminID *uint     `gorm:"index"`
	UserCode      string    `gorm:"unique"`

	Nickname     string
	DeviceID     string
	FCMToken     string
	SnsType      uint
	UserType     uint
	IsAgree      uint
	Images       []Image       `gorm:"foreignKey:Uid"`
	UserAfcs     []UserAfc     `gorm:"foreignKey:Uid"`
	UserDisables []UserDisable `gorm:"foreignKey:UID"`
	UserVisits   []UserVisit   `gorm:"foreignKey:UID"`
}
