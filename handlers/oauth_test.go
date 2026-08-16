package handlers

import (
	"forum/db"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// --- helper: create test store ---
func newTestStore(t *testing.T) *db.Store {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	conn, err := db.InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitializeDatabase: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	return db.NewStore(conn)
}

//
// ------------------------------------------------------------
// LOGIN HANDLERS (simple redirect tests)
// ------------------------------------------------------------
//

func TestGoogleLoginHandler_Redirect(t *testing.T) {
	store := newTestStore(t)

	req := httptest.NewRequest("GET", "/auth/google/login", nil)
	w := httptest.NewRecorder()

	GoogleLoginHandler(store).ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", w.Code)
	}

	if w.Header().Get("Location") == "" {
		t.Fatalf("expected Location header")
	}
}

func TestGitHubLoginHandler_Redirect(t *testing.T) {
	store := newTestStore(t)

	req := httptest.NewRequest("GET", "/auth/github/login", nil)
	w := httptest.NewRecorder()

	GitHubLoginHandler(store).ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", w.Code)
	}

	if w.Header().Get("Location") == "" {
		t.Fatalf("expected Location header")
	}
}

func TestDiscordLoginHandler_Redirect(t *testing.T) {
	store := newTestStore(t)

	req := httptest.NewRequest("GET", "/auth/discord/login", nil)
	w := httptest.NewRecorder()

	DiscordLoginHandler(store).ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", w.Code)
	}

	if w.Header().Get("Location") == "" {
		t.Fatalf("expected Location header")
	}
}

//
// ------------------------------------------------------------
// CALLBACK HANDLERS — missing code (pure logic)
// ------------------------------------------------------------
//

func TestGoogleCallbackHandler_MissingCode(t *testing.T) {
	store := newTestStore(t)

	req := httptest.NewRequest("GET", "/auth/google/callback", nil)
	w := httptest.NewRecorder()

	GoogleCallbackHandler(store).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGitHubCallbackHandler_MissingCode(t *testing.T) {
	store := newTestStore(t)

	req := httptest.NewRequest("GET", "/auth/github/callback", nil)
	w := httptest.NewRecorder()

	GitHubCallbackHandler(store).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDiscordCallbackHandler_MissingCode(t *testing.T) {
	store := newTestStore(t)

	req := httptest.NewRequest("GET", "/auth/discord/callback", nil)
	w := httptest.NewRecorder()

	DiscordCallbackHandler(store).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

//
// ------------------------------------------------------------
// CreateOrLinkOAuthUser (pure DB logic)
// ------------------------------------------------------------
//

func TestCreateOrLinkOAuthUser_NewUser(t *testing.T) {
	store := newTestStore(t)

	user, err := CreateOrLinkOAuthUser(store, "google", "id123", "g@example.com", "GUser")
	if err != nil {
		t.Fatalf("CreateOrLinkOAuthUser: %v", err)
	}

	if user.Email != "g@example.com" {
		t.Fatalf("email mismatch")
	}
	if user.OAuthProvider != "google" {
		t.Fatalf("provider mismatch")
	}
	if user.OAuthID != "id123" {
		t.Fatalf("oauthID mismatch")
	}
}

func TestCreateOrLinkOAuthUser_LinkExisting(t *testing.T) {
	store := newTestStore(t)

	// existing user with same email
	uid, _ := store.CreateUser("x@example.com", "x", "hashpw")

	user, err := CreateOrLinkOAuthUser(store, "github", "ghid", "x@example.com", "GHUser")
	if err != nil {
		t.Fatalf("CreateOrLinkOAuthUser: %v", err)
	}

	// reload user from DB to get updated OAuth fields
	user, _ = store.GetUserByID(uid)

	if user.ID != uid {
		t.Fatalf("should link to existing user")
	}
	if user.OAuthProvider != "github" {
		t.Fatalf("provider mismatch: got %s", user.OAuthProvider)
	}
}

//
// ------------------------------------------------------------
// CreateSessionAndSetCookie (pure logic)
// ------------------------------------------------------------
//

func TestCreateSessionAndSetCookie(t *testing.T) {
	store := newTestStore(t)

	uid, _ := store.CreateUser("u@example.com", "u", "hashpw")
	user, _ := store.GetUserByID(uid)

	w := httptest.NewRecorder()

	err := CreateSessionAndSetCookie(w, store, user)
	if err != nil {
		t.Fatalf("CreateSessionAndSetCookie: %v", err)
	}

	// cookie must exist
	found := false
	for _, c := range w.Result().Cookies() {
		if c.Name == "session" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected session cookie")
	}
}
