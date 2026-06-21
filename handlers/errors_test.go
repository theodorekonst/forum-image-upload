package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRenderErrorSetsStatusAndBody(t *testing.T) {
	recorder := httptest.NewRecorder()

	RenderError(recorder, http.StatusNotFound, "The page was not found.")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusNotFound)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "404") {
		t.Fatalf("body does not contain status code: %q", body)
	}
	if !strings.Contains(body, "The page was not found.") {
		t.Fatalf("body does not contain message: %q", body)
	}
}

func TestRenderErrorUsesDefaultMessage(t *testing.T) {
	recorder := httptest.NewRecorder()

	RenderError(recorder, http.StatusInternalServerError, "")

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(recorder.Body.String(), "Internal Server Error") {
		t.Fatalf("body does not contain default message: %q", recorder.Body.String())
	}
}
