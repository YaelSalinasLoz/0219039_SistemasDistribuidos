package main

import (
	"encoding/json" // Encode / Decode Json
	"fmt"           // Print output
	"log"           // logging errors
	"net/http"      // HTTP server
	"sync"          // Synchronization --> Mutex
)

// Structure of a user
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Global Variables
var users []User  // Slice --> store users
var nxID int      // ID Counter
var mu sync.Mutex // mUTEX --> Thread-safe access manager

// Handler --> add new users
func newUser(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	var tempUser User
	// Unmarshal
	err := json.NewDecoder(r.Body).Decode(&tempUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	nxID++
	tempUser.ID = nxID
	users = append(users, tempUser)
	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tempUser)
}

// Handler to get users list
func getUser(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	// Give user list
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Main
func main() {
	// Initialize variables
	nxID = 0
	users = []User{}

	// Set up route handlers
	http.HandleFunc("/addUser", newUser)
	http.HandleFunc("/getUser", getUser)
	// Start the server on port 8080
	fmt.Println("Server status: running on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
