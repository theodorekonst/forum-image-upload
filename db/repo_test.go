package db

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStoreCRUD(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, err := InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	defer conn.Close()

	store := NewStore(conn)

	// Create user
	userID, err := store.CreateUser("u1@example.com", "u1", "hashpw")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	u, err := store.GetUserByEmail("u1@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if u.ID != userID {
		t.Fatalf("user id mismatch: got %d want %d", u.ID, userID)
	}

	// Category
	catID, err := store.CreateCategory("TestCat")
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	cats, err := store.ListCategories()
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	if len(cats) == 0 {
		t.Fatalf("expected at least one category")
	}

	// Post
	postID, err := store.CreatePost(userID, "Hello", "World")
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}

	p, err := store.GetPostByID(postID)
	if err != nil {
		t.Fatalf("GetPostByID: %v", err)
	}
	if p.Title != "Hello" {
		t.Fatalf("post title mismatch: %s", p.Title)
	}

	// Assign category
	if err := store.AddPostCategory(postID, catID); err != nil {
		t.Fatalf("AddPostCategory: %v", err)
	}

	// Comment
	cID, err := store.CreateComment(postID, userID, "Nice post")
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}
	if cID == 0 {
		t.Fatalf("invalid comment id")
	}

	comments, err := store.ListCommentsByPost(postID)
	if err != nil {
		t.Fatalf("ListCommentsByPost: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment got %d", len(comments))
	}

	// Likes
	if _, err := store.AddLikeToPost(postID, userID); err != nil {
		t.Fatalf("AddLikeToPost: %v", err)
	}
	count, err := store.CountPostLikes(postID)
	if err != nil {
		t.Fatalf("CountPostLikes: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 like got %d", count)
	}

	// Sessions
	token := "tok123"
	expires := time.Now().Add(time.Hour)
	if _, err := store.CreateSession(userID, token, expires); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	sess, err := store.GetSession(token)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if sess.UserID != userID {
		t.Fatalf("session user mismatch: %d", sess.UserID)
	}

	if err := store.DeleteSession(token); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
}
