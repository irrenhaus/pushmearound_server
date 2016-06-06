package models

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID             uint
	CreatedAt      time.Time
	LastModifiedAt time.Time
	LastSignInAt   time.Time
	Username       string
	FirstName      string
	LastName       string
	Email          string
	EmailConfirmed bool
	Password       string
	Devices        []Device
	Tokens         []AccessToken
	Messages       []Message
	Friends        []Friendship
	FriendOf       []Friendship
}

type Friendship struct {
	ID             uint
	CreatedAt      time.Time
	LastModifiedAt time.Time
	User           User
	UserID         uint
	HasFriend      User
	HasFriendID    uint
}

func scanUser(row *sql.Row) (User, error) {
	u := User{}
	err := row.Scan(&u.ID, &u.CreatedAt, &u.LastModifiedAt, &u.Username, &u.FirstName, &u.LastName, &u.Email, &u.EmailConfirmed, &u.Password, &u.LastSignInAt)

	return u, err
}

func FindUser(DB *sql.DB, id uint) (User, error) {
	row := DB.QueryRow("SELECT * FROM users WHERE id=$1", id)

	return scanUser(row)
}

func FindUserByLogin(DB *sql.DB, login string) (User, error) {
	row := DB.QueryRow("SELECT * FROM users WHERE username=$1 OR email=$1", login)

	return scanUser(row)
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
