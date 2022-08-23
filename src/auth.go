package main

import (
	"golang.org/x/crypto/bcrypt"
	"regexp"
)

const emailRegex = "^(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])$"

func getLoginError(login string) string {
	if len(login) < 3 {
		return "Login must be at least 3 characters long"
	} else if len(login) > 60 {
		return "Login must be at most 60 characters long"
	} else if match, _ := regexp.MatchString("^[\\d\\w\\-_]+$", login); !match {
		return "Login may contain only alphanum chars, underscores and minus"
	} else {
		return ""
	}
}

func getEmailError(email string) string {
	if len(email) < 7 {
		return "Email must be at least 7 characters long"
	} else if len(email) > 150 {
		return "Email must be at most 150 characters long"
	} else if match, _ := regexp.MatchString(emailRegex, email); !match {
		return "Email is invalid"
	} else {
		return ""
	}
}

func getPasswordError(password string) string {
	if len(password) < 12 {
		return "Password must be at least 12 characters long"
	} else if len(password) > 60 {
		return "Password must be at most 60 characters long"
	} else {
		return ""
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
