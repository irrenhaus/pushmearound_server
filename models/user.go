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
	Messages       []Message    `gorm:"ForeignKey:FromUser"`
	Friends        []Friendship `gorm:"ForeignKey:User"`
	FriendOf       []Friendship `gorm:"ForeignKey:HasFriend"`
	LastSignInAt   time.Time
}

type Friendship struct {
	gorm.Model
	User        User
	UserID      uint
	HasFriend   User
	HasFriendID uint
}

func (u *User) FriendsWith(other *User) bool {
	if other.ID == u.ID {
		return true
	}

	for _, friend := range u.Friends {
		if friend.ID == other.ID {
			return true
		}
	}

	for _, friendOf := range u.FriendOf {
		if friendOf.ID == other.ID {
			return true
		}
	}

	return false
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
