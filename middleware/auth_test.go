package middleware

import (
	"forum/db"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

// TestGetUserFromRequest_NoCookie -> should return error when no session cookie exists
func TestGetUserFromRequest_NoCookie(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	conn, err := db.InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	defer conn.Close()

	store := db.NewStore(conn)

	req := httptest.NewRequest("GET", "/", nil)
	_, err = GetUserFromRequest(store, req)
	if err == nil {
		t.Fatalf("expected error when cookie is missing")
	}
}

// TestGetUserFromRequest_SessionNotFound -> cookie exists but DB has no session
func TestGetUserFromRequest_SessionNotFound(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	conn, err := db.InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	defer conn.Close()

	store := db.NewStore(conn)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "missing"})

	_, err = GetUserFromRequest(store, req)
	if err != db.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// TestGetUserFromRequest_ExpiredSession -> expired session should be deleted and return ErrNotFound
func TestGetUserFromRequest_ExpiredSession(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	conn, err := db.InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	defer conn.Close()

	store := db.NewStore(conn)

	uid, _ := store.CreateUser("u@example.com", "u", "hashpw")

	// Create expired session
	token := "expired123"
	expires := time.Now().Add(-1 * time.Hour)
	store.CreateSession(uid, token, expires)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})

	_, err = GetUserFromRequest(store, req)
	if err != db.ErrNotFound {
		t.Fatalf("expected ErrNotFound for expired session, got %v", err)
	}

	// Ensure session was deleted
	_, err = store.GetSession(token)
	if err != db.ErrNotFound {
		t.Fatalf("expected session to be deleted, got %v", err)
	}
}

// TestGetUserFromRequest_Valid -> should return correct user
func TestGetUserFromRequest_Valid(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	conn, err := db.InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	defer conn.Close()

	store := db.NewStore(conn)

	uid, _ := store.CreateUser("u@example.com", "u", "hashpw")

	token := "valid123"
	expires := time.Now().Add(1 * time.Hour)
	store.CreateSession(uid, token, expires)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})

	user, err := GetUserFromRequest(store, req)
	if err != nil {
		t.Fatalf("GetUserFromRequest: %v", err)
	}
	if user.ID != uid {
		t.Fatalf("user mismatch: got %d want %d", user.ID, uid)
	}
}

// TestRequireAuth_NoSession -> should redirect to home page
func TestRequireAuth_NoSession(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := RequireAuth(store, next)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", w.Code)
	}
	if w.Header().Get("Location") != "/" {
		t.Fatalf("expected redirect to '/', got %q", w.Header().Get("Location"))
	}
	if called {
		t.Fatalf("next handler should NOT be called")
	}
}

// TestRequireAuth_ValidSession -> should call next handler
func TestRequireAuth_ValidSession(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	uid, _ := store.CreateUser("u@example.com", "u", "hashpw")
	token := "auth123"
	expires := time.Now().Add(1 * time.Hour)
	store.CreateSession(uid, token, expires)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := RequireAuth(store, next)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Fatalf("expected next handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
}

// TestRequireAuth_ExpiredSession -> should redirect to home page, not login
func TestRequireAuth_ExpiredSession(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	uid, _ := store.CreateUser("u@example.com", "u", "hashpw")
	token := "expired-auth"
	expires := time.Now().Add(-1 * time.Hour)
	store.CreateSession(uid, token, expires)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := RequireAuth(store, next)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", w.Code)
	}
	if w.Header().Get("Location") != "/" {
		t.Fatalf("expected redirect to '/', got %q", w.Header().Get("Location"))
	}
	if called {
		t.Fatalf("next handler should NOT be called")
	}
}

// TestRequireGuest_LoggedIn -> should redirect to /
func TestRequireGuest_LoggedIn(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	uid, _ := store.CreateUser("u@example.com", "u", "hashpw")
	token := "guest123"
	expires := time.Now().Add(1 * time.Hour)
	store.CreateSession(uid, token, expires)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := RequireGuest(store, next)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", w.Code)
	}
	if called {
		t.Fatalf("next handler should NOT be called")
	}
}

// TestRequireGuest_Guest -> should call next handler
func TestRequireGuest_Guest(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := RequireGuest(store, next)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Fatalf("expected next handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
}
