package main

import (
	"encoding/json"
	"fmt"
	"log"

	"golang-chat/wrapper"

	socketio "github.com/googollee/go-socket.io"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Message represents a chat message
type Message struct {
	gorm.Model
	Text   string `json:"text"`
	Sender string `json:"sender"`
	Room   string `json:"room"`
}

func main() {
	// Create an Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Enable CORS for all routes
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))

	// Connect to SQLite database
	db, err := gorm.Open(sqlite.Open("chat.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	// Auto-migrate the models
	err = db.AutoMigrate(&Message{})
	if err != nil {
		log.Fatal("Error migrating models:", err)
	}

	// Set up routes
	e.GET("/", func(c echo.Context) error {
		return c.File("index.html")
	})

	e.Any("/socket.io/", socketIOWrapper(db))

	// Start the server
	e.Logger.Fatal(e.Start(":3000"))
}

func socketIOWrapper(db *gorm.DB) func(context echo.Context) error {
	server, err := wrapper.NewWrapperSocketIO(nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	server.OnConnect("/", func(context echo.Context, conn socketio.Conn) error {
		conn.SetContext("")
		fmt.Println("connected:", conn.ID())
		return nil
	})

	server.OnError("/", func(context echo.Context, e error) {
		fmt.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(context echo.Context, conn socketio.Conn, msg string) {
		fmt.Println("closed", msg)
	})

	server.OnEvent("/", "joinRoom", func(context echo.Context, conn socketio.Conn, inputRoom interface{}) {
		room := fmt.Sprintf("%v", inputRoom)
		fmt.Println("join:", room)

		conn.SetContext(room)
		conn.Join(room)

		// Get all messages for the room from the database
		var messages []Message
		err := db.Where("room = ?", room).Find(&messages).Error
		if err != nil {
			log.Println("Error retrieving messages:", err)
			return
		}

		// Emit the messages to the client
		conn.Emit("messages", messages)
	})

	server.OnEvent("/", "newMessage", func(context echo.Context, conn socketio.Conn, inputData interface{}) {
		// Convert inputData to data struct
		inputBytes, err := json.Marshal(inputData)
		if err != nil {
			fmt.Println("Error converting inputData to JSON:", err)
			return
		}

		var data Message
		err = json.Unmarshal(inputBytes, &data)
		if err != nil {
			fmt.Println("Error converting JSON to Message struct:", err)
			return
		}

		// Insert the new message into the database
		err = db.Create(&data).Error
		if err != nil {
			fmt.Println("Error saving message:", err)
			return
		}

		// Broadcast the new message to all clients in the room
		server.BroadcastToRoom("/", data.Room, "newMessage", data)
	})

	return server.HandlerFunc
}
