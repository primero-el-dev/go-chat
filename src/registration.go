package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
	"log"
)

type RegistrationForm struct {
	Login            string
	Email            string
	Password         string
	RepeatedPassword string
}

func (form *RegistrationForm) GetErrors(db *sql.DB) map[string]string {
	formErrors := make(map[string]string)

	if err := getRegistrationLoginError(db, form.Login); err != "" {
		formErrors["login"] = err
	}

	if err := getRegistrationEmailError(db, form.Email); err != "" {
		formErrors["email"] = err
	}

	if err := getPasswordError(form.Password); err != "" {
		formErrors["password"] = err
	} else if form.Password != form.RepeatedPassword {
		formErrors["password"] = "Both passwords must be the same"
	}

	return formErrors
}

func getRegistrationLoginError(db *sql.DB, login string) string {
	if err := getLoginError(login); err != "" {
		return err
	}

	rows, err := db.Query("SELECT * FROM app_user WHERE login = $1 LIMIT 1", login)
	if err != nil {
		fmt.Print(err.Error())
		return "An error occured. Please try again"
	}

	defer rows.Close()
	if rows.Next() {
		return "Given login is already taken. Please choose another one"
	}

	return ""
}

func getRegistrationEmailError(db *sql.DB, email string) string {
	if err := getEmailError(email); err != "" {
		return err
	}

	rows, err := db.Query("SELECT * FROM app_user WHERE email = $1 LIMIT 1", email)
	if err != nil {
		fmt.Print(err.Error())
		return "An error occured. Please try again"
	}

	defer rows.Close()
	if rows.Next() {
		return "Given email is already taken. Please choose another one"
	}

	return ""
}

func RegistrationGetHandler(ctx *fiber.Ctx, db *sql.DB) error {
	return ctx.Render("registration", fiber.Map{})
}

func RegistrationPostHandler(ctx *fiber.Ctx, db *sql.DB) error {
	form := RegistrationForm{}
	if err := ctx.BodyParser(&form); err != nil {
		log.Printf("An error occured: %v", err)
		return ctx.SendString(err.Error())
	}

	if db.Ping() != nil {
		return errors.New("Cannot connect database")
	}

	formErrors := form.GetErrors(db)
	if len(formErrors) != 0 {
		return ctx.Render("registration", fiber.Map{
			"Login":         form.Login,
			"Email":         form.Email,
			"LoginError":    formErrors["login"],
			"EmailError":    formErrors["email"],
			"PasswordError": formErrors["password"],
		})
	}

	hash, err := HashPassword(form.Password)
	if err != nil {
		log.Fatalf("An error occured while hashing password: %v", err)
	}

	_, err = db.Exec("INSERT INTO app_user (login, email, password) VALUES ($1, $2, $3)", form.Login, form.Email, hash)
	if err != nil {
		log.Fatalf("An error occured while executing query: %v", err)
	}

	return ctx.Redirect("/login")
}
