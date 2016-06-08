package models

import (
	"database/sql"
	"errors"
	"time"
)

type AccessToken struct {
	ID        string
	CreatedAt time.Time
	UserID    uint
}

func FindAccessToken(DB *sql.DB, token string) (AccessToken, error) {
	a := AccessToken{}
	err := DB.QueryRow("SELECT * FROM access_tokens WHERE id=$1", token).Scan(&a.ID, &a.CreatedAt, &a.UserID)

	return a, err
}

func (a *AccessToken) Create(DB *sql.DB) error {
	return DB.QueryRow("INSERT INTO access_tokens (id, created_at, user_id) VALUES ($1, current_timestamp(), $2) RETURNING id, created_at", a.ID, a.UserID).Scan(&a.ID, &a.CreatedAt)
}

func (a *AccessToken) Delete(DB *sql.DB) error {
	if a.ID == "" {
		return errors.New("AccessToken object has no ID")
	}

	res, err := DB.Exec("DELETE FROM access_tokens WHERE id=$1", a.ID)

	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected <= 0 {
		return errors.New("No such database entry")
	}

	return nil
}
