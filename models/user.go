package models

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
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
	UserID         uint
	HasFriendID    uint
}

func scanUser(row *sql.Row) (User, error) {
	u := User{}
	err := row.Scan(&u.ID, &u.CreatedAt, &u.LastModifiedAt, &u.LastSignInAt, &u.Username, &u.FirstName, &u.LastName, &u.Email, &u.EmailConfirmed, &u.Password)

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

func (u *User) Create(DB *sql.DB) error {
	return DB.QueryRow("INSERT INTO users (username, first_name, last_name, email, password) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, last_modified_at, last_sign_in_at, email_confirmed", u.Username, u.FirstName, u.LastName, u.Email, u.Password).Scan(&u.ID, &u.CreatedAt, &u.LastModifiedAt, &u.LastSignInAt, &u.EmailConfirmed)
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

func (u *User) LoadDevices(DB *sql.DB) error {
	var err error
	u.Devices, err = FindDevicesByUserID(DB, u.ID)
	return err
}
