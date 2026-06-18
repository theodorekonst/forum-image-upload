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

	// 2. Protect home page
	http.HandleFunc("/", middleware.RequireAuth(store, handlers.HomeHandler)) // HomeHandler -> HomePage (when home page will be created)

	// 3. Auth routes
	http.HandleFunc("/register", handlers.RegisterHandler(store))
	http.HandleFunc("/login", handlers.LoginHandler(store))
	http.HandleFunc("/logout", handlers.LogoutHandler(store))

	// 4. Start server
	log.Println("Server running on port :8080")
	http.ListenAndServe(":8080", nil)
}
