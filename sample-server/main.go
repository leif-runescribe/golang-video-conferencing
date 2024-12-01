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

// Server structure to encapsulate application state
type Server struct {
	Rooms        map[string]*Room      // Map of room IDs to Room objects
	Users        map[string]string     // Map of session IDs to user names
	RoomMutex    sync.Mutex            // Mutex for room map
	UserMutex    sync.Mutex            // Mutex for user map
	SessionStore *sessions.CookieStore // Session store
}

// Room structure
type Room struct {
	ID      string                     `json:"id"`
	Members map[string]*websocket.Conn // Member names and connections
	Lock    sync.Mutex                 // Mutex for room members
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Create a new Server instance
func NewServer() *Server {
	return &Server{
		Rooms:        make(map[string]*Room),
		Users:        make(map[string]string),
		SessionStore: sessions.NewCookieStore([]byte("secret-key")),
	}
}

// Handle user registration
func (s *Server) registerUser(w http.ResponseWriter, r *http.Request) {
	// Ensure the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve the session
	session, err := s.SessionStore.Get(r, "user-session")
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
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || reqBody.Name == "" {
		http.Error(w, "Invalid name", http.StatusBadRequest)
		return
	}

	// Store the name in the session
	session.Values["name"] = reqBody.Name
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Add user to server's user map
	s.UserMutex.Lock()
	s.Users[session.ID] = reqBody.Name
	s.UserMutex.Unlock()

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
		"name":    reqBody.Name,
	})
}

// Create a new room
func (s *Server) createRoom(w http.ResponseWriter, r *http.Request) {
	// Get user name from session
	session, _ := s.SessionStore.Get(r, "user-session")
	userName, ok := session.Values["name"].(string)
	if !ok || userName == "" {
		http.Error(w, "User not registered", http.StatusUnauthorized)
		return
	}

	// Generate room ID and create room
	roomID := s.generateRoomID()
	room := &Room{
		ID:      roomID,
		Members: make(map[string]*websocket.Conn),
	}

	// Add room to server
	s.RoomMutex.Lock()
	s.Rooms[roomID] = room
	s.RoomMutex.Unlock()

	// Respond with room ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"roomID": roomID,
		"name":   userName,
	})
}

// Generate a unique room ID
func (s *Server) generateRoomID() string {
	return fmt.Sprintf("room-%d", len(s.Rooms)+1)
}

// Join a room
func (s *Server) joinRoom(w http.ResponseWriter, r *http.Request) {
	// Get user name from session
	session, _ := s.SessionStore.Get(r, "user-session")
	userName, ok := session.Values["name"].(string)
	if !ok || userName == "" {
		http.Error(w, "User not registered", http.StatusUnauthorized)
		return
	}

	// Get room ID from URL params
	vars := mux.Vars(r)
	roomID := vars["roomID"]

	// Check if room exists
	s.RoomMutex.Lock()
	room, exists := s.Rooms[roomID]
	s.RoomMutex.Unlock()

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

	// Add user to the room
	room.Lock.Lock()
	room.Members[userName] = conn
	room.Lock.Unlock()

	// Notify others that the user joined
	s.broadcastMessage(room, fmt.Sprintf("%s joined the room!", userName))

	log.Printf("User %s joined Room %s\n", userName, roomID)

	// Handle messages
	s.handleMessages(room, userName, conn)
}

// Handle incoming WebSocket messages
func (s *Server) handleMessages(room *Room, userName string, conn *websocket.Conn) {
	defer func() {
		// Remove user from the room
		room.Lock.Lock()
		delete(room.Members, userName)
		room.Lock.Unlock()

		s.broadcastMessage(room, fmt.Sprintf("%s left the room!", userName))
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}
		log.Printf("Message from %s in room %s: %s\n", userName, room.ID, string(message))
		s.broadcastMessage(room, fmt.Sprintf("%s: %s", userName, message))
	}
}

// Broadcast a message to all members in a room
func (s *Server) broadcastMessage(room *Room, message string) {
	room.Lock.Lock()
	defer room.Lock.Unlock()

	for _, conn := range room.Members {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("Error broadcasting message:", err)
		}
	}
}

// List all registered users
func (s *Server) listRegisteredUsers(w http.ResponseWriter, r *http.Request) {
	s.UserMutex.Lock()
	defer s.UserMutex.Unlock()

	users := make([]string, 0, len(s.Users))
	for _, name := range s.Users {
		users = append(users, name)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{
		"users": users,
	})
}
func (s *Server) goofy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{
		"msg": "sup",
	})
}

// Main function
func main() {
	server := NewServer()

	r := mux.NewRouter()
	r.HandleFunc("/register", server.registerUser).Methods("POST")
	r.HandleFunc("/create-room", server.createRoom).Methods("POST")
	r.HandleFunc("/join-room/{roomID}", server.joinRoom).Methods("GET")
	r.HandleFunc("/list-users", server.listRegisteredUsers).Methods("GET")
	r.HandleFunc("/", goofy).Methods("GET")
	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", r)
}
