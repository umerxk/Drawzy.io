package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"skribbl-clone/config"
	"skribbl-clone/db"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func printAllCollections() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collections, err := db.MessagesCollection1.Database().ListCollectionNames(ctx, struct{}{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	fmt.Println("Collections in the database:")
	for _, collection := range collections {
		fmt.Println(collection)
	}
	return nil
}
func SaveMessageToMongo(msg Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define the document to insert
	document := bson.M{
		"room":      msg.Room,
		"username":  msg.Username,
		"text":      msg.Text,
		"timestamp": time.Now(), // Optional: add creation timestamp
	}

	log.Println("Saving message to MongoDB:", document)

	_, err := db.MessagesCollection1.InsertOne(ctx, document)
	if err != nil {
		return fmt.Errorf("failed to save message to MongoDB: %w", err)
	}

	return nil
}

// GetMessagesForRoom fetches messages for a specific room
func GetMongoMessagesForRoom(room string) ([]Message, error) {
	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define the filter for the query
	filter := bson.M{"room": room}

	// Define sort options (sort by timestamp in ascending order)
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "timestamp", Value: 1}})

	// Query the collection
	cursor, err := db.MessagesCollection1.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}
	defer cursor.Close(ctx)

	// Parse the results
	var messages []Message
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, fmt.Errorf("error decoding messages: %w", err)
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

	messages, err := GetMongoMessagesForRoom(room)
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

		err := SaveMessageToMongo(msg)
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
	db.InitMongoDB()
	go handleMessages()
	err2 := printAllCollections()
	if err2 != nil {
		fmt.Println("Error:", err2)
	}
	fmt.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Error starting server: " + err.Error())
	}
}
