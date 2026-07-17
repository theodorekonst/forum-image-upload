package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCreatePostHandlerRejectsWhitespaceOnlyInput(t *testing.T) {
	form := url.Values{
		"title":   {"   \t"},
		"content": {"\n   "},
	}
	req := httptest.NewRequest(http.MethodPost, "/create-post", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	CreatePostHandler(nil)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for whitespace-only post, got %d", w.Code)
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

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for whitespace-only comment, got %d", w.Code)
	}
}
