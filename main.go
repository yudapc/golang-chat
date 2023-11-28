package main

import (
	"log"

	"golang-chat/internal/message"

	pkgsocketio "golang-chat/pkg/socketio"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Message represents a chat message

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
	err = db.AutoMigrate(&message.Message{})
	if err != nil {
		log.Fatal("Error migrating models:", err)
	}

	// Set up routes
	e.GET("/", func(c echo.Context) error {
		return c.File("index.html")
	})

	e.Any("/socket.io/", pkgsocketio.SocketIOHandler(db))

	// Start the server
	e.Logger.Fatal(e.Start(":3000"))
}
