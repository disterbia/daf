package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model       // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 자동으로 추가
	Name             string
	Nickname         string
	Email            string    `gorm:"unique"`
	Birthday         time.Time `gorm:"type:date"`
	Phone            string    `gorm:"unique"`
	DeviceID         string    `json:"device_ID"`
	Gender           bool      // true:남 false: 여
	FCMToken         string    `json:"fcm_token"`
	SnsType          uint
	Images           []Image           `gorm:"foreignKey:Uid"`
	UserJointActions []UserJointAction `gorm:"foreignKey:Uid"`
}
