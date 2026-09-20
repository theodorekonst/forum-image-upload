package handlers

import (
	"forum/db"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCreatePostHandlerRejectsWhitespaceOnlyInput(t *testing.T) {
	dir := t.TempDir()
	conn, err := db.InitializeDatabase(dir + "/test.db")
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	defer conn.Close()
	store := db.NewStore(conn)

	form := url.Values{
		"title":   {"   \t"},
		"content": {"\n   "},
	}
	req := httptest.NewRequest(http.MethodPost, "/create-post", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	CreatePostHandler(store)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for whitespace-only post, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Title cannot be empty.") || !strings.Contains(w.Body.String(), "Content cannot be empty.") {
		t.Fatalf("expected inline validation error")
	}
}

func TestCreateCommentHandlerRejectsWhitespaceOnlyInput(t *testing.T) {
	form := url.Values{
		"post_id": {"1"},
		"content": {" \n\t "},
	}
	req := httptest.NewRequest(http.MethodPost, "/comment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	CreateCommentHandler(nil)(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect for whitespace-only comment, got %d", w.Code)
	}
	if location := w.Header().Get("Location"); location != "/post?id=1&comment_error=required&scroll=comment-form" {
		t.Fatalf("unexpected redirect location: %s", location)
	}
}

func TestCreatePostHandlerRequiresCategory(t *testing.T) {
	conn, err := db.InitializeDatabase(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	defer conn.Close()

	form := url.Values{"title": {"Valid title"}, "content": {"Valid content"}}
	req := httptest.NewRequest(http.MethodPost, "/post/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	CreatePostHandler(db.NewStore(conn))(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when category is missing, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Choose at least one category.") {
		t.Fatalf("expected category validation message")
	}
}

func TestViewPostHandlerRejectsPostMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/post?id=1", nil)
	w := httptest.NewRecorder()

	ViewPostHandler(nil)(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestCreateCommentHandlerReturnsNotFoundForMissingPost(t *testing.T) {
	conn, err := db.InitializeDatabase(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	defer conn.Close()

	form := url.Values{"post_id": {"999999"}, "content": {"Valid comment"}}
	req := httptest.NewRequest(http.MethodPost, "/comment/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	CreateCommentHandler(db.NewStore(conn))(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPostReactHandlerReturnsNotFoundForMissingPost(t *testing.T) {
	conn, err := db.InitializeDatabase(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	defer conn.Close()

	form := url.Values{"post_id": {"999999"}, "reaction_type": {"like"}}
	req := httptest.NewRequest(http.MethodPost, "/post/react", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	PostReactHandler(db.NewStore(conn))(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
