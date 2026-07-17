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
	if !strings.Contains(w.Body.String(), "Title and content cannot be empty.") {
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
