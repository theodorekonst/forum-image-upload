package main

import (
	"forum/db"
	"forum/handlers"
	"forum/middleware"
	"log"
	"net/http"
)

func main() {

	// 1. Initialise database
	connection, err := db.InitializeDatabase("forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	store := db.NewStore(connection)

	// Create custom mux
	mux := http.NewServeMux()

	// 2. Protect home page
	mux.Handle("/", middleware.RequireAuth(store, http.HandlerFunc(handlers.HomeHandler))) // HomeHandler -> HomePage (when created)

	// 3. Guest-only routes
	mux.Handle("/register", middleware.RequireGuest(store, handlers.RegisterHandler(store)))
	mux.Handle("/login", middleware.RequireGuest(store, handlers.LoginHandler(store)))

	// 4. Logout (requires auth)
	mux.Handle("/logout", middleware.RequireAuth(store, handlers.LogoutHandler(store)))

	// 5. Start server
	log.Println("Server running on port :8080")
	http.ListenAndServe(":8080", mux)
}
