package models

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
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
	Tokens         []AccessToken
	LastSignInAt   time.Time
}

func (u *User) SetPassword(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	u.Password = string(hash)
	return nil
}

func (u *User) ComparePassword(plaintextPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plaintextPassword))
}
