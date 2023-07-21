package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	red   = "\033[31m"
	green = "\033[32m"
	blue  = "\033[34m"
	reset = "\033[0m"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var clients = make(map[string]*websocket.Conn)
var voiceConnections = make(map[string]*websocket.Conn) // Map to store voice chat connections

var mutex sync.Mutex


func main() {
	fmt.Println(green + "Starting....")
	http.HandleFunc("/ws", handleWebSocket)

	// Start the HTTP server
	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection to WebSocket:", err)
		return
	}
	defer conn.Close()

	for {
		// Read the message from the WebSocket connection
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			// Handle disconnection or error
			handleDisconnection(conn)
			log.Println("Failed to read message from WebSocket:", err)
			break
		}

		// Deserialize the received message into ChatDto object
		chat := ChatDto{}
		err = json.Unmarshal(message, &chat)
		if err != nil {
			log.Println("Failed to deserialize message:", err)
			continue
		}
		//log.Printf("Received message: %+v", chat)
		switchMessage(chat,messageType,conn)
	}
}

func switchMessage(message ChatDto,messageType int,conn *websocket.Conn){
	
	switch message.Type {
	case "newConnection":
		// Code to execute when expression equals value1
		// Register the client connection
		clients[message.From] = conn
		log.Printf("Client connected: %+v", message.From)
	case "whoisonline":
		// Get the list of online users
		sendOnlineUsers(conn)
	case "chat":
		recipientConn, ok := clients[message.To]
		if !ok {
			log.Printf("Recipient not found: %s", message.To)
		}

		// Send the message to the recipient
		err := recipientConn.WriteMessage(messageType, marshalMessage(message))
		if err != nil {
			log.Println("Failed to send message to recipient:", err)
			break
		}
	case "voiceRequest":
		recipientConn, ok := clients[message.To]
		if !ok {
			log.Printf("Recipient not found: %s", message.To)
			break
		}

		// Store the voice connection for both the sender and recipient
		voiceConnections[message.From] = conn
		voiceConnections[message.To] = recipientConn

		// Send a voice chat request to the recipient
		err := recipientConn.WriteMessage(messageType, marshalMessage(message))
		if err != nil {
			log.Println("Failed to send voice chat request to recipient:", err)
			break
		}
	case "voiceAccept":
		recipientConn, ok := voiceConnections[message.From]
		if !ok {
			log.Printf("Voice connection not found for: %s", message.From)
			break
		}

		// Send a voice chat acceptance to the sender
		err := recipientConn.WriteMessage(messageType, marshalMessage(message))
		if err != nil {
			log.Println("Failed to send voice chat acceptance to sender:", err)
			break
		}	
	default:
		// Code to execute when none of the above cases match
		fmt.Println(red + "Unknown Message Type: " + message.Type)
	}
}

func marshalMessage(message ChatDto) []byte{
	jsonData, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Marshalling the message broke... ")
	}
	return jsonData

}

func getOnlineUsers() []string {
	mutex.Lock()
	defer mutex.Unlock()
	users := make([]string, 0, len(clients))
	for user := range clients {
		users = append(users, user)
	}
	return users
}

func sendOnlineUsers(conn *websocket.Conn) {
	users := getOnlineUsers()
	fmt.Println("Users online:")
	fmt.Println(users)
	// Create an online users message
	message := ChatDto{
		From : "server",
		Type: "whoisonline",
		Users: users,
		Timestamp: time.Now().Unix(),
	}

	// Serialize the message to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Println("Failed to serialize online users message:", err)
		return
	}

	// Send the online users message to the client
	err = conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		log.Println("Failed to send online users message to client:", err)
		return
	}
}

func handleDisconnection(conn *websocket.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	// Find the client's username from the clients map
	var username string
	for user, c := range clients {
		if c == conn {
			username = user
			break
		}
	}

	// Remove the disconnected client from the clients map
	delete(clients, username)

	log.Println("Client disconnected:", username)
}

type ChatDto struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Msg       string `json:"message"`
	Type      string `json:"type"`
	Users []string `json:"users"`
	Timestamp int64  `json:"timestamp"`
}

