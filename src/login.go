package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"log"
)

type LoginForm struct {
	Login    string
	Password string
}

func (form *LoginForm) GetErrors(db *sql.DB) map[string]string {
	formErrors := make(map[string]string)

	if err := getExistingLoginError(db, form.Login); err != "" {
		formErrors["login"] = "Wrong credentials"
	}

	if err := getPasswordError(form.Password); err != "" {
		formErrors["login"] = "Wrong credentials"
	}

	return formErrors
}

func getExistingLoginError(db *sql.DB, login string) string {
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

func LoginGetHandler(ctx *fiber.Ctx, db *sql.DB) error {
	return ctx.Render("login", fiber.Map{})
}

func LoginPostHandler(sess *session.Session, ctx *fiber.Ctx, db *sql.DB) error {
	form := LoginForm{}
	if err := ctx.BodyParser(&form); err != nil {
		log.Printf("An error occured: %v", err)
		return ctx.SendString(err.Error())
	}

	if db.Ping() != nil {
		return errors.New("Cannot connect database")
	}

	// Validation
	errorMessage := "Invalid credentials"

	renderFormWithError := func(errorMsg string) error {
		return ctx.Render("login", fiber.Map{
			"Login": form.Login,
			"Error": errorMsg,
		})
	}

	if getLoginError(form.Login) != "" {
		return renderFormWithError(errorMessage)
	}

	if getPasswordError(form.Password) != "" {
		return renderFormWithError(errorMessage)
	}

	rows, err := db.Query("SELECT * FROM app_user WHERE login = $1 LIMIT 1", form.Login)
	if err != nil {
		log.Fatalf("An error occured while executing query: %v", err)
		return renderFormWithError("An error occured. Please try again")
	}

	defer rows.Close()
	if !rows.Next() {
		return renderFormWithError(errorMessage)
	}

	user := User{}
	if rows.Scan(&user.Id, &user.Login, &user.Email, &user.Password, &user.CreatedAt) != nil {
		log.Fatalf("An error occured while executing query: %v", err)
		return renderFormWithError("An error occured. Please try again")
	}

	if !CheckPasswordHash(form.Password, user.Password) {
		return renderFormWithError(errorMessage)
	}

	// Login user
	sess.Set(userIdKey, user.Id)

	if err := sess.Save(); err != nil {
		panic(err)
	}

	ctx.SendStatus(200)

	return ctx.Redirect("/room/1")
}
