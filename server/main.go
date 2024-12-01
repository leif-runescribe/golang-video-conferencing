package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)

// Session store
var store = sessions.NewCookieStore([]byte("secret-key"))

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

// Handle user registration (getting their name)
func registerUser(w http.ResponseWriter, r *http.Request) {
	// Ensure the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve the session
	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Unable to retrieve session", http.StatusInternalServerError)
		return
	}

	// Check if user already has a name in the session
	if userName, ok := session.Values["name"].(string); ok && userName != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "User already registered",
			"name":    userName,
		})
		return
	}

	// Parse and validate the JSON request body
	var reqBody struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	// Validate name (should not be empty or just whitespace)
	if reqBody.Name == "" {
		http.Error(w, "Invalid name provided", http.StatusBadRequest)
		return
	}

	// Store the name in the session
	session.Values["name"] = reqBody.Name
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
		"name":    reqBody.Name,
	})
}

// Create Room
func createRoom(w http.ResponseWriter, r *http.Request) {
	// Get user name from session
	session, _ := store.Get(r, "user-session")
	userName, ok := session.Values["name"].(string)
	if !ok || userName == "" {
		http.Error(w, "User not registered", http.StatusUnauthorized)
		return
	}

	// Generate room ID and create room
	roomID := generateRoomID()
	room := &Room{
		ID:      roomID,
		Members: make(map[string]*websocket.Conn),
	}

	// Add room to global map
	roomMutex.Lock()
	rooms[roomID] = room
	roomMutex.Unlock()

	// Respond with room ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"roomID": roomID,
		"name":   userName,
	})
}

func test(w http.ResponseWriter, r *http.Request) string {

	fmt.Println("yoooman")
	return "hi"
}

// Generate a simple room ID (can be improved with UUID)
func generateRoomID() string {
	return fmt.Sprintf("room-%d", len(rooms)+1)
}

// Join Room
func joinRoom(w http.ResponseWriter, r *http.Request) {
	// Get user name from session
	session, _ := store.Get(r, "user-session")
	userName, ok := session.Values["name"].(string)
	if !ok || userName == "" {
		http.Error(w, "User not registered", http.StatusUnauthorized)
		return
	}

	// Get room ID from URL params
	vars := mux.Vars(r)
	roomID := vars["roomID"]

	// Lock rooms map and check if room exists
	roomMutex.Lock()
	room, exists := rooms[roomID]
	roomMutex.Unlock()

	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	// Add member to the room
	room.Lock.Lock()
	room.Members[userName] = conn
	room.Lock.Unlock()

	// Notify others that the user joined
	broadcastMessage(room, fmt.Sprintf("%s joined the room!", userName))

	log.Printf("User %s joined Room %s\n", userName, roomID)

	// Handle incoming messages from this user
	handleMessages(room, userName, conn)
}

// Handle WebSocket messages from a user
func handleMessages(room *Room, userName string, conn *websocket.Conn) {
	defer func() {
		// Remove user from the room on exit
		room.Lock.Lock()
		delete(room.Members, userName)
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
		log.Printf("Message from %s in room %s: %s\n", userName, room.ID, string(message))
		broadcastMessage(room, fmt.Sprintf("%s: %s", userName, message))
	}
}

// Broadcast a message to all members in the room
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

// List all members in a room
func listMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomID"]

	roomMutex.Lock()
	room, exists := rooms[roomID]
	roomMutex.Unlock()

	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// List member names
	room.Lock.Lock()
	defer room.Lock.Unlock()

	memberNames := make([]string, 0, len(room.Members))
	for memberName := range room.Members {
		memberNames = append(memberNames, memberName)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{
		"members": memberNames,
	})
}

// Simple test handler
func testPostHandler(w http.ResponseWriter, r *http.Request) {
	// Log request method and path
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)

	// Check Content-Type
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	// Read request body
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Failed to parse body: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Request body: %v", body)

	// Respond back with the same body
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Data received successfully",
		"data":    body,
	})
}

func main() {
	// Set up router
	r := mux.NewRouter()

	// Routes
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Request: %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	r.HandleFunc("/register", registerUser).Methods("POST")
	r.HandleFunc("/create-room", createRoom).Methods("POST")
	r.HandleFunc("/join-room/{roomID}", joinRoom).Methods("GET")
	r.HandleFunc("/list-members/{roomID}", listMembers).Methods("GET")
	r.HandleFunc("/test-post", testPostHandler).Methods("POST")

	// Start server
	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", r)
}
