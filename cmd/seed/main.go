package main

import (
	"fmt"
	"log"

	"forum/db"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	const dbPath = "forum.db"

	conn, err := db.InitializeDatabase(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer conn.Close()

	store := db.NewStore(conn)

	categories := []string{"Programming", "Sports", "Gaming", "Movies"}
	for _, name := range categories {
		if _, err := store.CreateCategory(name); err != nil {
			log.Printf("create category %s: %v", name, err)
		}
	}

	// create a test user
	email := "test@example.com"
	username := "member1"
	password := "password123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}
	if _, err := store.CreateUser(email, username, string(hash)); err != nil {
		log.Printf("create user: %v", err)
	}

	user, err := store.GetUserByEmail(email)
	if err == nil {
		postID, err := store.CreatePost(user.ID, "Welcome to the forum", "This is the first seeded post.")
		if err != nil {
			log.Printf("create post: %v", err)
		} else {
			// assign first category if present
			if err := store.AddPostCategory(postID, 1); err != nil {
				log.Printf("assign category: %v", err)
			}
		}
	}

	fmt.Println("Seeding complete")
}
