package pkgsocketio

import (
	"encoding/json"
	"fmt"
	"log"

	"golang-chat/internal/message"

	socketio "github.com/googollee/go-socket.io"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Implementation Handler
func SocketIOHandler(db *gorm.DB) func(context echo.Context) error {
	server, err := NewWrapperSocketIO(nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	namespace := "/"

	server.OnConnect(namespace, func(context echo.Context, conn socketio.Conn) error {
		conn.SetContext("")
		fmt.Println("connected:", conn.ID())
		return nil
	})

	server.OnError(namespace, func(context echo.Context, e error) {
		fmt.Println("meet error:", e)
	})

	server.OnDisconnect(namespace, func(context echo.Context, conn socketio.Conn, msg string) {
		fmt.Println("closed", msg)
	})

	server.OnEvent(namespace, "joinRoom", func(context echo.Context, conn socketio.Conn, inputRoom interface{}) {
		room := fmt.Sprintf("%v", inputRoom)
		fmt.Println("join:", room)

		conn.SetContext(room)
		conn.Join(room)

		// Get all messages for the room from the database
		var messages []message.Message
		err := db.Where("room = ?", room).Find(&messages).Error
		if err != nil {
			log.Println("Error retrieving messages:", err)
			return
		}

		// Emit the messages to the client
		conn.Emit("messages", messages)
	})

	server.OnEvent(namespace, "newMessage", func(context echo.Context, conn socketio.Conn, inputData interface{}) {
		// Convert inputData to data struct
		inputBytes, err := json.Marshal(inputData)
		if err != nil {
			fmt.Println("Error converting inputData to JSON:", err)
			return
		}

		var data message.Message
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
		server.BroadcastToRoom(namespace, data.Room, "newMessage", data)
	})

	return server.HandlerFunc
}
