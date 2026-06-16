package main

import (
	"forum/db"
	"forum/handlers"
	"log"
	"net/http"
)

func main() {

	// 1. Initialise database
	connection, err := db.InitializeDatabase("forum.db")
	if err != nil {
		log.Fatal(err)
	}
	store := db.NewStore(connection)

	// 2. Register routes
	http.HandleFunc("/register", handlers.RegisterHandler(store))
	http.HandleFunc("/login", handlers.LoginHandler(store))

	// 3. Start server
	log.Println("Server running on port :8080")
	http.ListenAndServe(":8080", nil)
}
