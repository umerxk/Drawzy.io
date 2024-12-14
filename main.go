package main

import (
	"fmt"
	"net/http"
	"skribbl-clone/config"
	"skribbl-clone/db"

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

func saveMessageToDB(msg Message) error {
	query := `INSERT INTO messages (room, username, text) VALUES ($1, $2, $3)`
	_, err := db.DB.Exec(query, msg.Room, msg.Username, msg.Text)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	return nil
}

func getMessagesForRoom(room string) ([]Message, error) {
	query := `SELECT room, username, text FROM messages WHERE room = $1 ORDER BY created_at ASC`
	rows, err := db.DB.Query(query, room)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.Room, &msg.Username, &msg.Text)
		if err != nil {
			return nil, fmt.Errorf("error scanning message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
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

	messages, err := getMessagesForRoom(room)
	if err == nil {
		for _, msg := range messages {
			ws.WriteJSON(msg)
		}
	}

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
		fmt.Println("printing message to DB:", msg)

		err := saveMessageToDB(msg)
		if err != nil {
			fmt.Println("Error saving message to DB:", err)
			continue
		}

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
	config.LoadConfig()
	db.InitDB()
	go handleMessages()

	fmt.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Error starting server: " + err.Error())
	}
}
