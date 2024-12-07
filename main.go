// package main

// import (
// 	"log"
// 	"skribbl-clone/config"
// 	"skribbl-clone/db"
// 	"skribbl-clone/routes"

// 	"github.com/gofiber/fiber/v2"
// )

// func main() {
// 	// Initialize Fiber
// 	app := fiber.New()

// 	// Load configuration
// 	config.LoadConfig()
// 	db.InitDB()

// 	// Initialize routes
// 	routes.SetupRoutes(app)

// 	// Start the server
// 	log.Println("Starting server on :3000")
// 	if err := app.Listen(":3000"); err != nil {
// 		log.Fatalf("Error starting server: %v", err)
// 	}
// }

package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var broadcast = make(chan Message)
var rooms = make(map[string]map[*websocket.Conn]bool)

type Message struct {
	Room     string `json:"room"`     // Target room
	Username string `json:"username"` // Username of the sender
	Text     string `json:"text"`     // Message content
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to websocket:", err)
		return
	}
	defer ws.Close()

	var joinRequest struct {
		Type string `json:"type"`
		Room string `json:"room"`
	}

	err = ws.ReadJSON(&joinRequest) // Parse the join request
	if err != nil || joinRequest.Type != "join" {
		fmt.Println("Error reading join request or invalid type:", err)
		return
	}

	room := joinRequest.Room

	// Add client to the specified room
	if rooms[room] == nil {
		rooms[room] = make(map[*websocket.Conn]bool)
	}
	rooms[room][ws] = true

	defer func() {
		delete(rooms[room], ws) // Remove client when they disconnect
		if len(rooms[room]) == 0 {
			delete(rooms, room) // Clean up empty room
		}
	}()

	// Read messages from the client
	for {
		var msg Message
		fmt.Printf("Broadcasting message: %+v\n", msg)
		err := ws.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading JSON:", err)
			break
		}

		msg.Room = room // Ensure the message is tagged with the correct room
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast

		// Broadcast only to clients in the specified room
		if clientsInRoom, ok := rooms[msg.Room]; ok {
			for client := range clientsInRoom {
				err := client.WriteJSON(msg)
				if err != nil {
					fmt.Println("Error writing JSON:", err)
					client.Close()
					delete(clientsInRoom, client)
				}
			}
		}
	}
}

func main() {
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	fmt.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Error starting server: " + err.Error())
	}
}
