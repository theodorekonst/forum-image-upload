package handlers

import (
	"forum/db"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func deleteTestStore(t *testing.T) *db.Store {
	t.Helper()
	conn, err := db.InitializeDatabase(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return db.NewStore(conn)
}

func authenticatedDeleteRequest(t *testing.T, store *db.Store, userID, postID int64) *http.Request {
	t.Helper()
	token := "delete-test-token-" + strconv.FormatInt(userID, 10)
	if _, err := store.CreateSession(userID, token, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	form := url.Values{"post_id": {strconv.FormatInt(postID, 10)}}
	req := httptest.NewRequest(http.MethodPost, "/post/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	return req
}

func authenticatedCommentDeleteRequest(t *testing.T, store *db.Store, userID, commentID int64) *http.Request {
	t.Helper()
	token := "comment-delete-token-" + strconv.FormatInt(userID, 10)
	if _, err := store.CreateSession(userID, token, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	form := url.Values{"comment_id": {strconv.FormatInt(commentID, 10)}}
	req := httptest.NewRequest(http.MethodPost, "/comment/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	return req
}

func TestDeletePostHandlerAllowsOwner(t *testing.T) {
	store := deleteTestStore(t)
	userID, _ := store.CreateUser("owner@example.com", "owner", "hash")
	postID, _ := store.CreatePost(userID, "Delete me", "content")
	req := authenticatedDeleteRequest(t, store, userID, postID)
	w := httptest.NewRecorder()

	DeletePostHandler(store)(w, req)

	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/" {
		t.Fatalf("expected redirect home, got status %d and location %q", w.Code, w.Header().Get("Location"))
	}
	if _, err := store.GetPostByID(postID); err != db.ErrNotFound {
		t.Fatalf("expected deleted post, got %v", err)
	}
}

func TestDeletePostHandlerRejectsNonOwner(t *testing.T) {
	store := deleteTestStore(t)
	ownerID, _ := store.CreateUser("owner@example.com", "owner", "hash")
	otherID, _ := store.CreateUser("other@example.com", "other", "hash")
	postID, _ := store.CreatePost(ownerID, "Keep me", "content")
	req := authenticatedDeleteRequest(t, store, otherID, postID)
	w := httptest.NewRecorder()

	DeletePostHandler(store)(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
	if _, err := store.GetPostByID(postID); err != nil {
		t.Fatalf("post should still exist: %v", err)
	}
}

func TestDeleteCommentHandlerAllowsOwner(t *testing.T) {
	store := deleteTestStore(t)
	userID, _ := store.CreateUser("owner@example.com", "owner", "hash")
	postID, _ := store.CreatePost(userID, "Post", "content")
	commentID, _ := store.CreateComment(postID, userID, "Delete me")
	req := authenticatedCommentDeleteRequest(t, store, userID, commentID)
	w := httptest.NewRecorder()

	DeleteCommentHandler(store)(w, req)

	location := "/post?id=" + strconv.FormatInt(postID, 10)
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != location {
		t.Fatalf("expected redirect to post, got status %d and location %q", w.Code, w.Header().Get("Location"))
	}
	if _, err := store.GetCommentByID(commentID); err != db.ErrNotFound {
		t.Fatalf("expected deleted comment, got %v", err)
	}
}

func TestDeleteCommentHandlerRejectsNonOwner(t *testing.T) {
	store := deleteTestStore(t)
	ownerID, _ := store.CreateUser("owner@example.com", "owner", "hash")
	otherID, _ := store.CreateUser("other@example.com", "other", "hash")
	postID, _ := store.CreatePost(ownerID, "Post", "content")
	commentID, _ := store.CreateComment(postID, ownerID, "Keep me")
	req := authenticatedCommentDeleteRequest(t, store, otherID, commentID)
	w := httptest.NewRecorder()

	DeleteCommentHandler(store)(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
	if _, err := store.GetCommentByID(commentID); err != nil {
		t.Fatalf("comment should still exist: %v", err)
	}
}
