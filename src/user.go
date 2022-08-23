package main

import (
	"database/sql"
	"time"
)

type User struct {
	Id        int
	Login     string
	Email     string
	Password  string
	CreatedAt time.Time
}

func (user *User) CreateFromRegistration(form *RegistrationForm) *User {
	user.Login = form.Login
	user.Email = form.Email
	user.Password = form.Password

	return user
}

func FetchUser(rows *sql.Rows) *User {
	if !rows.Next() {
		return nil
	}

	user := User{}
	err := rows.Scan(&user.Id, &user.Login, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		panic(err)
	}

	return &user
}
