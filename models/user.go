package models

import (
	"github.com/irrenhaus/pushmearound_server/models"
	"github.com/jinzhu/gorm"
	"time"
)

type User struct {
	gorm.Model
	Username       string `gorm:"unique_index"`
	FirstName      string
	LastName       string
	Email          string
	EmailConfirmed bool
	Password       string
	Devices        []Device
	LastSignInAt   time.Time
}
