package models

import (
	"errors"
	"github.com/jinzhu/gorm"
)

// Use a map so that we can check for existence without iterating over an array
var DevicePlatforms map[string]bool = map[string]bool{
	"chrome":  true,
	"firefox": true,
	"android": true,
}

type DeviceOptions struct {
	ID                uint `gorm:"primary_key"`
	DeviceID          uint `gorm:"index"`
	PushNotifications bool
}

type Device struct {
	gorm.Model
	UserID           uint   `gorm:"index"`
	DeviceID         string `gorm:"unique_index"`
	Platform         string `gorm:"index"`
	Options          DeviceOptions
	SentMessages     []Message `gorm:"ForeignKey:FromDevice"`
	ReceivedMessages []Message
}

func (d *DeviceOptions) ParseJSONMap(json map[string]interface{}) error {
	deviceID, ok := json["device_id"]
	if !ok {
		return errors.New("No device ID specified")
	}

	d.DeviceID, ok = deviceID.(uint)
	if !ok {
		return errors.New("Device ID is not an unsigned integer")
	}

	pn, ok := json["push_notifications"]
	if ok {
		if d.PushNotifications, ok = pn.(bool); !ok {
			return errors.New("Push notifications value is not a boolean")
		}
	}

	return nil
}
