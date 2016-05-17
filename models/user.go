package models

var userSchema = `
CREATE TABLE user (
	username text,
	first_name text,
	last_name text,
	email text
);`

type User struct {
	Username  string
	FirstName string `db:first_name`
	LastName  string `db:last_name`
	Email     string
}
