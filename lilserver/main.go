package main

import (
	"fmt"
	"lilserve/utils"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Room structure
type Room struct {
	ID      string                     `json:"id"`
	Members map[string]*websocket.Conn // Member names and connections
	Lock    sync.Mutex                 // Mutex for concurrency
}

// In-memory storage of rooms
var rooms = make(map[string]*Room)
var roomMutex sync.Mutex

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Generate a simple room ID (UUID for uniqueness)
func generateRoomID() string {
	return utils.GenId()
}

// Register a new user (Generates a unique user ID)
func registerUser(c *gin.Context) {
	// Parse the JSON request body
	var reqBody struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil || reqBody.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid name provided"})
		return
	}

	// Generate a unique user ID
	userID := utils.GenId()

	// Respond with the user details
	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
		"userID":  userID,
		"name":    reqBody.Name,
	})
}

// Create a new room
func createRoom(c *gin.Context) {
	// Parse the JSON request body
	var reqBody struct {
		UserID string `json:"userID" binding:"required"`
		Name   string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil || reqBody.UserID == "" || reqBody.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user or name provided"})
		return
	}

	// Create a new room with a unique ID
	roomID := generateRoomID()
	room := &Room{
		ID:      roomID,
		Members: make(map[string]*websocket.Conn),
	}

	// Add the room to the global map
	roomMutex.Lock()
	rooms[roomID] = room
	roomMutex.Unlock()

	// Respond with the room details
	c.JSON(http.StatusOK, gin.H{
		"message": "Room created successfully",
		"roomID":  roomID,
	})
}

// Join a room
func joinRoom(c *gin.Context) {
	roomID := c.Param("roomID")
	userID := c.Query("userID")
	userName := c.Query("name")

	if roomID == "" || userID == "" || userName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing room ID, user ID, or name"})
		return
	}

	// Check if the room exists
	roomMutex.Lock()
	room, exists := rooms[roomID]
	roomMutex.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Add the user to the room
	room.Lock.Lock()
	room.Members[userID] = conn
	room.Lock.Unlock()

	// Notify others that the user joined
	broadcastMessage(room, fmt.Sprintf("%s joined the room!", userName))

	log.Printf("User %s (%s) joined Room %s\n", userName, userID, roomID)

	// Handle incoming messages from the user
	handleMessages(room, userID, conn, userName)
}

// List members in a room
func listMembers(c *gin.Context) {
	roomID := c.Param("roomID")

	// Check if the room exists
	roomMutex.Lock()
	room, exists := rooms[roomID]
	roomMutex.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Collect all member IDs
	room.Lock.Lock()
	defer room.Lock.Unlock()

	memberNames := make([]string, 0, len(room.Members))
	for userID := range room.Members {
		memberNames = append(memberNames, userID)
	}

	c.JSON(http.StatusOK, gin.H{"members": memberNames})
}

// Handle WebSocket messages from a user
func handleMessages(room *Room, userID string, conn *websocket.Conn, userName string) {
	defer func() {
		// Remove user from the room on exit
		room.Lock.Lock()
		delete(room.Members, userID)
		room.Lock.Unlock()
		broadcastMessage(room, fmt.Sprintf("%s left the room!", userName))
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		log.Printf("Message from %s (%s) in room %s: %s\n", userName, userID, room.ID, string(message))
		broadcastMessage(room, fmt.Sprintf("%s: %s", userName, string(message)))
	}
}

// Broadcast a message to all members in a room
func broadcastMessage(room *Room, message string) {
	room.Lock.Lock()
	defer room.Lock.Unlock()

	for _, conn := range room.Members {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("Error broadcasting message:", err)
		}
	}
}

func main() {
	router := gin.Default()

	// User registration
	router.POST("/register", registerUser)

	// Room operations
	router.POST("/create-room", createRoom)
	router.GET("/join-room/:roomID", joinRoom)
	router.GET("/list-members/:roomID", listMembers)

	// Start the server
	log.Println("Server is running on http://localhost:8080")
	router.Run(":8080")
}
