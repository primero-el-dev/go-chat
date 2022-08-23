package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"math/rand"
	"time"
)

type Token struct {
	Id      int
	Value   string
	ValidTo time.Time
	RoomId  int
	UserId  int
}

func (token *Token) Init(roomId, userId int) *Token {
	token.Value = RandString(80)
	token.ValidTo = *generateTokenExpiry()
	token.RoomId = roomId
	token.UserId = userId

	return token
}

func (token *Token) Insert(db *sql.DB) *Token {
	_, err := db.Exec(
		"INSERT INTO token (value, valid_to, room_id, user_id) VALUES ($1, $2, $3, $4) RETURNING id",
		token.Value,
		token.ValidTo,
		token.RoomId,
		token.UserId)
	if err != nil {
		panic(err)
	}

	return token.FindLast(db)
}

func (token *Token) SetAsOnly(db *sql.DB) *Token {
	token.DeleteOtherForRoomAndUser(db)

	return token.Insert(db)
}

func (token *Token) DeleteOtherForRoomAndUser(db *sql.DB) {
	_, err := db.Exec("DELETE FROM token WHERE room_id = $1 AND user_id = $2", token.RoomId, token.UserId)
	if err != nil {
		panic(err)
	}
}

func (token *Token) DeleteOld(db *sql.DB) {
	_, err := db.Exec("DELETE FROM token WHERE valid_to < NOW()")
	if err != nil {
		panic(err)
	}
}

func (token *Token) DeleteForUser(db *sql.DB, user User) {
	_, err := db.Exec("DELETE FROM token WHERE user_id = $1", user.Id)
	if err != nil {
		panic(err)
	}
}

func (token *Token) FindValidByValue(db *sql.DB, value string) *Token {
	rows, err := db.Query("SELECT * FROM token WHERE valid_to >= NOW() AND value = $1", value)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	return FetchToken(rows)
}

func (token *Token) FindLast(db *sql.DB) *Token {
	rows, err := db.Query("SELECT * FROM token ORDER BY id DESC LIMIT 1")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	return FetchToken(rows)
}

func (token *Token) FindOwner(db *sql.DB, value string) *User {
	rows, err := db.Query("SELECT app_user.* FROM app_user INNER JOIN token ON token.user_id = app_user.id WHERE token.value = $1 LIMIT 1", value)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	return FetchUser(rows)
}

func FetchToken(rows *sql.Rows) *Token {
	if !rows.Next() {
		return nil
	}

	token := Token{}
	err := rows.Scan(&token.Id, &token.Value, &token.ValidTo, &token.RoomId, &token.UserId)
	if err != nil {
		panic(err)
	}

	return &token
}

func generateTokenExpiry() *time.Time {
	expiry := time.
		Now().
		Local().
		Add(time.Minute * time.Duration(20))

	return &expiry
}

func InitRandom() {
	rand.Seed(time.Now().UnixNano())
}

func RandString(n int) string {
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)

	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}

	return string(b)
}
