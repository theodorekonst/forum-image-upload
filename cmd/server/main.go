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

	// 2. Public home page and static frontend assets
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.Handle("/", handlers.HomeHandler(store))

	// 3. Guest-only routes
	mux.Handle("/register", middleware.RequireGuest(store, handlers.RegisterHandler(store)))
	mux.Handle("/login", middleware.RequireGuest(store, handlers.LoginHandler(store)))

	// 4. Logout (requires auth)
	mux.Handle("/logout", middleware.RequireAuth(store, handlers.LogoutHandler(store)))

	mux.Handle("/post/create", middleware.RequireAuth(store, handlers.CreatePostHandler(store)))
	mux.Handle("/post/delete", middleware.RequireAuth(store, handlers.DeletePostHandler(store)))
	mux.Handle("/post/react", middleware.RequireAuth(store, handlers.PostReactHandler(store)))
	mux.Handle("/post", handlers.ViewPostHandler(store))
	mux.Handle("/comment/create", middleware.RequireAuth(store, handlers.CreateCommentHandler(store)))
	mux.Handle("/comment/delete", middleware.RequireAuth(store, handlers.DeleteCommentHandler(store)))
	mux.Handle("/comment/react", middleware.RequireAuth(store, handlers.CommentReactHandler(store)))

	// 5. Start server
	log.Println("Server running on port :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
