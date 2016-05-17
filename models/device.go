package models

import (
	"github.com/jinzhu/gorm"
)

type Device struct {
	gorm.Model
	UserID            int    `gorm:"index"`
	DeviceId          string `gorm:"unique_index"`
	Platform          int    `gorm:"index"`
	PushNotifications bool
}
