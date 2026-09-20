package db

import (
	"errors"
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

	// Post reactions
	if err := store.SetPostReaction(postID, userID, "like"); err != nil {
		t.Fatalf("SetPostReaction like: %v", err)
	}
	count, err := store.CountPostLikes(postID)
	if err != nil {
		t.Fatalf("CountPostLikes: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 like got %d", count)
	}

	// Switch to dislike
	if err := store.SetPostReaction(postID, userID, "dislike"); err != nil {
		t.Fatalf("SetPostReaction dislike: %v", err)
	}
	count, err = store.CountPostLikes(postID)
	if err != nil {
		t.Fatalf("CountPostLikes after switch: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 likes after switch got %d", count)
	}
	dislikes, err := store.CountPostDislikes(postID)
	if err != nil {
		t.Fatalf("CountPostDislikes: %v", err)
	}
	if dislikes != 1 {
		t.Fatalf("expected 1 dislike got %d", dislikes)
	}

	// Remove reaction (toggle off)
	if err := store.RemovePostReaction(postID, userID); err != nil {
		t.Fatalf("RemovePostReaction: %v", err)
	}
	dislikes, err = store.CountPostDislikes(postID)
	if err != nil {
		t.Fatalf("CountPostDislikes after remove: %v", err)
	}
	if dislikes != 0 {
		t.Fatalf("expected 0 dislikes after remove got %d", dislikes)
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

func TestDeletePost(t *testing.T) {
	// 1. Temporary database
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	conn, err := InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	defer conn.Close()

	store := NewStore(conn)

	// 2. Create a user because every post needs an author
	userID, err := store.CreateUser(
		"delete@example.com",
		"deleteuser",
		"hashpw",
	)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// 3. Create the post we are going to delete
	postID, err := store.CreatePost(
		userID,
		"Post to delete",
		"This post will be deleted",
	)
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	categoryID, _ := store.CreateCategory("Delete cascade")
	if err := store.AddPostCategory(postID, categoryID); err != nil {
		t.Fatalf("AddPostCategory: %v", err)
	}
	commentID, _ := store.CreateComment(postID, userID, "Delete with post")
	if err := store.SetPostReaction(postID, userID, "like"); err != nil {
		t.Fatalf("SetPostReaction: %v", err)
	}
	if err := store.SetCommentReaction(commentID, userID, "like"); err != nil {
		t.Fatalf("SetCommentReaction: %v", err)
	}

	// 4. Delete it
	if err := store.DeletePost(postID); err != nil {
		t.Fatalf("DeletePost: %v", err)
	}

	// 5. Verify that it no longer exists
	_, err = store.GetPostByID(postID)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := store.GetCommentByID(commentID); err != ErrNotFound {
		t.Fatalf("expected comments to cascade, got %v", err)
	}
	var categoryLinks, postReactions, commentReactions int
	store.DB.QueryRow(`SELECT COUNT(*) FROM post_categories WHERE post_id = ?`, postID).Scan(&categoryLinks)
	store.DB.QueryRow(`SELECT COUNT(*) FROM post_likes WHERE post_id = ?`, postID).Scan(&postReactions)
	store.DB.QueryRow(`SELECT COUNT(*) FROM comment_likes WHERE comment_id = ?`, commentID).Scan(&commentReactions)
	if categoryLinks != 0 || postReactions != 0 || commentReactions != 0 {
		t.Fatalf("expected related rows to cascade, got categories=%d post reactions=%d comment reactions=%d", categoryLinks, postReactions, commentReactions)
	}
}

func TestDeleteCommentUsesTombstoneWhenRepliesExist(t *testing.T) {
	dir := t.TempDir()
	conn, err := InitializeDatabase(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	defer conn.Close()
	store := NewStore(conn)

	userID, _ := store.CreateUser("comments@example.com", "comments", "hash")
	postID, _ := store.CreatePost(userID, "Post", "Content")
	parentID, _ := store.CreateCommentWithParent(postID, userID, nil, "Parent")
	replyID, _ := store.CreateCommentWithParent(postID, userID, &parentID, "Reply")

	softDeleted, err := store.DeleteComment(parentID)
	if err != nil {
		t.Fatalf("DeleteComment parent: %v", err)
	}
	if !softDeleted {
		t.Fatalf("expected parent comment to use a tombstone")
	}
	parent, err := store.GetCommentByID(parentID)
	if err != nil || !parent.IsDeleted || parent.Content != "" {
		t.Fatalf("expected deleted parent tombstone, got %+v, %v", parent, err)
	}
	if _, err := store.GetCommentByID(replyID); err != nil {
		t.Fatalf("reply should remain: %v", err)
	}

	softDeleted, err = store.DeleteComment(replyID)
	if err != nil {
		t.Fatalf("DeleteComment reply: %v", err)
	}
	if softDeleted {
		t.Fatalf("expected leaf reply to be hard deleted")
	}
	if _, err := store.GetCommentByID(replyID); err != ErrNotFound {
		t.Fatalf("expected reply to be deleted, got %v", err)
	}
	if _, err := store.GetCommentByID(parentID); err != ErrNotFound {
		t.Fatalf("expected empty parent tombstone to be cleaned up, got %v", err)
	}
}

func TestInitializeDatabaseCreatesDefaultCategories(t *testing.T) {
	conn, err := InitializeDatabase(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	defer conn.Close()

	categories, err := NewStore(conn).ListCategories()
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	if len(categories) < 5 {
		t.Fatalf("expected default categories, got %d", len(categories))
	}
}

func TestCreatePostWithCategoriesIsAtomic(t *testing.T) {
	conn, err := InitializeDatabase(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	defer conn.Close()
	store := NewStore(conn)

	userID, err := store.CreateUser("atomic@example.com", "atomic_user", "hash")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if _, err := store.CreatePostWithCategories(userID, "Title", "Content", []int64{999999}, nil, "", 0); !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("expected ErrInvalidCategory, got %v", err)
	}

	var count int
	if err := store.DB.QueryRow(`SELECT COUNT(*) FROM posts WHERE author_id = ?`, userID).Scan(&count); err != nil {
		t.Fatalf("count posts: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected transaction rollback, found %d posts", count)
	}
}

func TestCreatePostWithCategoriesRequiresCategory(t *testing.T) {
	conn, err := InitializeDatabase(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	defer conn.Close()
	store := NewStore(conn)

	userID, err := store.CreateUser("required@example.com", "required_user", "hash")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if _, err := store.CreatePostWithCategories(userID, "Title", "Content", nil, nil, "", 0); !errors.Is(err, ErrCategoryRequired) {
		t.Fatalf("expected ErrCategoryRequired, got %v", err)
	}
}
