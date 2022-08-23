package main

import "database/sql"

type Room struct {
	Id   int
	Name string
}

func (room *Room) Find(db *sql.DB, id int) *Room {
	rows, err := db.Query("SELECT * FROM room WHERE id = $1", id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	return FetchRoom(rows)
}

func (room *Room) FindAll(db *sql.DB) []Room {
	rows, err := db.Query("SELECT * FROM room")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var rooms []Room
	for {
		r := FetchRoom(rows)
		if r == nil {
			break
		}
		rooms = append(rooms, *r)
	}

	return rooms
}

func FetchRoom(rows *sql.Rows) *Room {
	if !rows.Next() {
		return nil
	}

	room := Room{}
	err := rows.Scan(&room.Id, &room.Name)
	if err != nil {
		panic(err)
	}

	return &room
}
