package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html"
	"github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strconv"
)

type Message struct {
	User    string `json:"username"`
	Content string `json:"content"`
}

type PassedMessageDto struct {
	Token   string
	Content string
}

type MessageDto struct {
	User    *User
	Content string
}

//var clients = make(map[*websocket.Conn]bool)
//var channels = make(map[string]chan Message)

const userIdKey = "user_id"

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	store := session.New()

	// Connect to database
	connStr := fmt.Sprintf(
		"user=%s password=%s host=%s dbname=%s sslmode=disable",
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_NAME"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error on opening database connection: %s", err.Error())
	}

	engine := html.New("../view", ".html")
	engine.Reload(true)
	engine.Debug(true)
	app := fiber.New(fiber.Config{
		Views: engine,
		//ViewsLayout: "../view/main.html",
	})
	app.Static("/", "../public")

	getLoggedUserId := func(ctx *fiber.Ctx) int {
		sess, err := store.Get(ctx)
		if err != nil {
			panic(err)
		}

		userId := sess.Get(userIdKey)
		if userId == nil {
			return 0
		}

		return userId.(int)
	}

	//var userTokens := make(map[string]*User)
	InitRandom()

	app.Get("/room/:roomId", func(ctx *fiber.Ctx) error {
		userId := getLoggedUserId(ctx)
		if userId == 0 {
			return ctx.Redirect("/login")
		}

		roomId, err := strconv.Atoi(ctx.Params("roomId"))
		if err != nil {
			return ctx.Redirect("/room/1")
		}

		room := &Room{}
		room = room.Find(db, roomId)
		if room == nil {
			return ctx.Redirect("/room/1")
		}
		rooms := room.FindAll(db)

		token := Token{}
		token.Init(roomId, userId)
		token.SetAsOnly(db)

		return ctx.Render("index", fiber.Map{
			"Rooms":      rooms,
			"Room":       room,
			"TokenValue": token.Value,
		})
	})

	app.Get("/register", func(ctx *fiber.Ctx) error {
		userId := getLoggedUserId(ctx)
		if userId != 0 {
			return ctx.Redirect("/room/1")
		}

		return ctx.Render("registration", fiber.Map{})
	})

	app.Post("/register", func(ctx *fiber.Ctx) error {
		userId := getLoggedUserId(ctx)
		if userId != 0 {
			return ctx.Redirect("/room/1")
		}

		return RegistrationPostHandler(ctx, db)
	})

	app.Get("/login", func(ctx *fiber.Ctx) error {
		userId := getLoggedUserId(ctx)
		if userId != 0 {
			return ctx.Redirect("/room/1")
		}

		return ctx.Render("login", fiber.Map{})
	})

	app.Post("/login", func(ctx *fiber.Ctx) error {
		sess, err := store.Get(ctx)
		if err != nil {
			panic(err)
		}

		userId := sess.Get(userIdKey)
		if userId != nil && userId != 0 {
			return ctx.Redirect("/room/1")
		}

		return LoginPostHandler(sess, ctx, db)
	})

	app.Get("/logout", func(ctx *fiber.Ctx) error {
		sess, err := store.Get(ctx)
		if err != nil {
			panic(err)
		}

		sess.Destroy()
		sess.Save()

		return ctx.Redirect("/login")
	})

	app.Use("/ws", func(ctx *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		_, err := store.Get(ctx)
		if err != nil {
			panic(err)
		}

		if websocket.IsWebSocketUpgrade(ctx) {
			ctx.Locals("allowed", true)
			return ctx.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	var roomConnectionDetails = make(map[*Room]*ConnectionDetails)

	room := &Room{}
	for _, room := range room.FindAll(db) {
		connectionDetails := ConnectionDetails{}
		connectionDetails.Init()
		roomConnectionDetails[&room] = &connectionDetails

		go runHub(&connectionDetails)

		app.Get(fmt.Sprintf("/ws/%d", room.Id), websocket.New(func(conn *websocket.Conn) {
			defer func() {
				connectionDetails.Unregister <- conn
				conn.Close()
			}()

			connectionDetails.Register <- conn

			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Println("read error: ", err)
					}

					return
				}

				passedMessageDto := PassedMessageDto{}
				token := Token{}
				err = json.Unmarshal(message, &passedMessageDto)
				if err != nil {
					return
				}
				user := token.FindOwner(db, passedMessageDto.Token)
				if user == nil {
					return
				}
				messageDto := MessageDto{User: user, Content: passedMessageDto.Content}

				if messageType == websocket.TextMessage {
					connectionDetails.Broadcast <- &messageDto
				} else {
					log.Println("websocket message received of type: ", messageType)
				}
			}
		}))
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatalln(app.Listen(fmt.Sprintf(":%v", port)))
}

type ConnectionDetails struct {
	Clients    map[*websocket.Conn]client
	Register   chan *websocket.Conn
	Broadcast  chan *MessageDto
	Unregister chan *websocket.Conn
}

func (connectionDetails *ConnectionDetails) Init() {
	connectionDetails.Clients = make(map[*websocket.Conn]client)
	connectionDetails.Register = make(chan *websocket.Conn)
	connectionDetails.Broadcast = make(chan *MessageDto)
	connectionDetails.Unregister = make(chan *websocket.Conn)
}

type client struct{}

func runHub(connectionDetails *ConnectionDetails) {
	for {
		select {
		case connection := <-connectionDetails.Register:
			connectionDetails.Clients[connection] = client{}
			log.Println("connection registered")

		case messageDto := <-connectionDetails.Broadcast:
			for connection := range connectionDetails.Clients {
				returnMessage := map[string]string{
					"login":   messageDto.User.Login,
					"content": messageDto.Content,
				}
				jsonString, err := json.Marshal(returnMessage)
				if err != nil {
					panic(err)
				}
				err = connection.WriteMessage(websocket.TextMessage, jsonString)
				if err != nil {
					connectionDetails.Unregister <- connection
					connection.WriteMessage(websocket.CloseMessage, []byte{})
					connection.Close()
				}
			}

		case connection := <-connectionDetails.Unregister:
			delete(connectionDetails.Clients, connection)
		}
	}
}

//func CreateChannel(name string) {
//	if _, exists := channels[name]; !exists {
//		channels[name] = make(chan Message)
//	}
//}
//
//func (msg *Message) Send(channelName string) {
//	if ch, ok := channels[channelName]; ok {
//		ch <- *msg
//	}
//}
