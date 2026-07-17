package db

import (
    "path/filepath"
    "testing"
    "time"
)


// TestCreateSession -> verifies session creation, timestamps, and stored fields
func TestCreateSession(t *testing.T) {
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "test.db")

    conn, err := InitializeDatabase(dbPath)
    if err != nil {
        t.Fatalf("InitializeDatabase: %v", err)
    }
    defer conn.Close()

    store := NewStore(conn)

    uid, err := store.CreateUser("u@example.com", "u", "hashpw")
    if err != nil {
        t.Fatalf("CreateUser: %v", err)
    }

    token := "tok123"
    expires := time.Now().Add(1 * time.Hour)

    id, err := store.CreateSession(uid, token, expires)
    if err != nil {
        t.Fatalf("CreateSession: %v", err)
    }
    if id == 0 {
        t.Fatalf("expected non-zero session ID")
    }

    sess, err := store.GetSession(token)
    if err != nil {
        t.Fatalf("GetSession: %v", err)
    }

    if sess.UserID != uid {
        t.Fatalf("user mismatch: got %d want %d", sess.UserID, uid)
    }
    if sess.Token != token {
        t.Fatalf("token mismatch: got %s want %s", sess.Token, token)
    }
    if !sess.ExpiresAt.Equal(expires) {
        t.Fatalf("expires_at mismatch: got %v want %v", sess.ExpiresAt, expires)
    }
    if sess.CreatedAt.IsZero() {
        t.Fatalf("created_at should not be zero")
    }
}


// TestGetSession_NotFound -> ensures unknown tokens return ErrNotFound
func TestGetSession_NotFound(t *testing.T) {
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "test.db")

    conn, err := InitializeDatabase(dbPath)
    if err != nil {
        t.Fatalf("InitializeDatabase: %v", err)
    }
    defer conn.Close()

    store := NewStore(conn)

    _, err = store.GetSession("does-not-exist")
    if err == nil {
        t.Fatalf("expected error for missing session")
    }
    if err != ErrNotFound {
        t.Fatalf("expected ErrNotFound, got %v", err)
    }
}


// TestDeleteSession -> verifies deletion and correct ErrNotFound behavior
func TestDeleteSession(t *testing.T) {
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "test.db")

    conn, err := InitializeDatabase(dbPath)
    if err != nil {
        t.Fatalf("InitializeDatabase: %v", err)
    }
    defer conn.Close()

    store := NewStore(conn)

    uid, _ := store.CreateUser("u@example.com", "u", "hashpw")
    token := "tok123"
    expires := time.Now().Add(time.Hour)

    _, err = store.CreateSession(uid, token, expires)
    if err != nil {
        t.Fatalf("CreateSession: %v", err)
    }

    if err := store.DeleteSession(token); err != nil {
        t.Fatalf("DeleteSession: %v", err)
    }

    err = store.DeleteSession(token)
    if err != ErrNotFound {
        t.Fatalf("expected ErrNotFound on second delete, got %v", err)
    }

    _, err = store.GetSession(token)
    if err != ErrNotFound {
        t.Fatalf("expected ErrNotFound after deletion, got %v", err)
    }
}


// TestDeleteSessionsByUserID -> ensures all sessions for a user are removed
func TestDeleteSessionsByUserID(t *testing.T) {
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "test.db")

    conn, err := InitializeDatabase(dbPath)
    if err != nil {
        t.Fatalf("InitializeDatabase: %v", err)
    }
    defer conn.Close()

    store := NewStore(conn)

    uid, _ := store.CreateUser("u@example.com", "u", "hashpw")

    expires := time.Now().Add(time.Hour)
    _, _ = store.CreateSession(uid, "tokA", expires)
    _, _ = store.CreateSession(uid, "tokB", expires)

    if err := store.DeleteSessionsByUserID(uid); err != nil {
        t.Fatalf("DeleteSessionsByUserID: %v", err)
    }

    _, err = store.GetSession("tokA")
    if err != ErrNotFound {
        t.Fatalf("expected tokA to be deleted, got %v", err)
    }
    _, err = store.GetSession("tokB")
    if err != ErrNotFound {
        t.Fatalf("expected tokB to be deleted, got %v", err)
    }
}


// TestSessionCascadeDelete -> ensures sessions are removed when user is deleted (FK cascade)
func TestSessionCascadeDelete(t *testing.T) {
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "test.db")

    conn, err := InitializeDatabase(dbPath)
    if err != nil {
        t.Fatalf("InitializeDatabase: %v", err)
    }
    defer conn.Close()

    store := NewStore(conn)

    uid, _ := store.CreateUser("u@example.com", "u", "hashpw")

    expires := time.Now().Add(time.Hour)
    _, _ = store.CreateSession(uid, "tok123", expires)

    _, err = store.DB.Exec(`DELETE FROM users WHERE id = ?`, uid)
    if err != nil {
        t.Fatalf("delete user: %v", err)
    }

    _, err = store.GetSession("tok123")
    if err != ErrNotFound {
        t.Fatalf("expected session to be deleted via cascade, got %v", err)
    }
}


// TestGetUserByID -> ensures user lookup by ID works and handles ErrNotFound
func TestGetUserByID(t *testing.T) {
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "test.db")

    conn, err := InitializeDatabase(dbPath)
    if err != nil {
        t.Fatalf("InitializeDatabase: %v", err)
    }
    defer conn.Close()

    store := NewStore(conn)

    uid, err := store.CreateUser("u@example.com", "u", "hashpw")
    if err != nil {
        t.Fatalf("CreateUser: %v", err)
    }

    u, err := store.GetUserByID(uid)
    if err != nil {
        t.Fatalf("GetUserByID: %v", err)
    }
    if u.ID != uid {
        t.Fatalf("ID mismatch: got %d want %d", u.ID, uid)
    }

    _, err = store.GetUserByID(9999)
    if err != ErrNotFound {
        t.Fatalf("expected ErrNotFound for missing user, got %v", err)
    }
}


// TestGetUserByUsername -> ensures lookup by username works and handles ErrNotFound
func TestGetUserByUsername(t *testing.T) {
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "test.db")

    conn, err := InitializeDatabase(dbPath)
    if err != nil {
        t.Fatalf("InitializeDatabase: %v", err)
    }
    defer conn.Close()

    store := NewStore(conn)

    uid, err := store.CreateUser("u@example.com", "u", "hashpw")
    if err != nil {
        t.Fatalf("CreateUser: %v", err)
    }

    u, err := store.GetUserByUsername("u")
    if err != nil {
        t.Fatalf("GetUserByUsername: %v", err)
    }
    if u.ID != uid {
        t.Fatalf("ID mismatch: got %d want %d", u.ID, uid)
    }

    _, err = store.GetUserByUsername("doesnotexist")
    if err != ErrNotFound {
        t.Fatalf("expected ErrNotFound for missing username, got %v", err)
    }
}


// TestGetUserByEmail -> ensures lookup by email works and handles ErrNotFound
func TestGetUserByEmail(t *testing.T) {
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "test.db")

    conn, err := InitializeDatabase(dbPath)
    if err != nil {
        t.Fatalf("InitializeDatabase: %v", err)
    }
    defer conn.Close()

    store := NewStore(conn)

    uid, err := store.CreateUser("u@example.com", "u", "hashpw")
    if err != nil {
        t.Fatalf("CreateUser: %v", err)
    }

    u, err := store.GetUserByEmail("u@example.com")
    if err != nil {
        t.Fatalf("GetUserByEmail: %v", err)
    }
    if u.ID != uid {
        t.Fatalf("ID mismatch: got %d want %d", u.ID, uid)
    }

    _, err = store.GetUserByEmail("missing@example.com")
    if err != ErrNotFound {
        t.Fatalf("expected ErrNotFound for missing email, got %v", err)
    }
}
