package models

import (
	"github.com/jinzhu/gorm"
)

const (
	ContentTypeMessage = iota
	ContentTypeURL     = iota
	ContentTypeFile    = iota
)

type Message struct {
	gorm.Model
	User         User
	UserID       uint `gorm:"index"`
	Device       Device
	DeviceID     uint `gorm:"index"`
	FromUser     User
	FromUserID   uint `gorm:"index"`
	FromDevice   Device
	FromDeviceID uint `gorm:"index"`
	ContentType  uint
	Title        string
	Text         string `gorm:"type:text"`
	URL          string
}
