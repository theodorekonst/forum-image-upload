package handlers

import (
	"forum/db"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// TestHomeHandler -> should return placeholder text
func TestHomeHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	HomeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Home Page Placeholder") {
		t.Fatalf("expected placeholder text, got %s", w.Body.String())
	}
}

// TestRegisterHandler_Get -> should return placeholder page
func TestRegisterHandler_Get(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	req := httptest.NewRequest("GET", "/register", nil)
	w := httptest.NewRecorder()

	RegisterHandler(store)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Register Page Placeholder") {
		t.Fatalf("expected placeholder text")
	}
}

// TestRegisterHandler_MissingFields -> should return 400
func TestRegisterHandler_MissingFields(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	req := httptest.NewRequest("POST", "/register", nil)
	req.Form = url.Values{
		"email":            {""},
		"username":         {""},
		"password":         {""},
		"confirm_password": {""},
	}
	w := httptest.NewRecorder()

	RegisterHandler(store)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// NEW TEST: confirm password mismatch
func TestRegisterHandler_ConfirmPasswordMismatch(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	req := httptest.NewRequest("POST", "/register", nil)
	req.Form = url.Values{
		"email":            {"test@example.com"},
		"username":         {"validuser"},
		"password":         {"Password1!"},
		"confirm_password": {"Different1!"},
	}
	w := httptest.NewRecorder()

	RegisterHandler(store)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for mismatched confirm password, got %d", w.Code)
	}
}

// TestRegisterHandler_InvalidEmail -> invalid email returns 400
func TestRegisterHandler_InvalidEmail(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	req := httptest.NewRequest("POST", "/register", nil)
	req.Form = url.Values{
		"email":            {"invalid"},
		"username":         {"validuser"},
		"password":         {"Password1!"},
		"confirm_password": {"Password1!"},
	}
	w := httptest.NewRecorder()

	RegisterHandler(store)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestRegisterHandler_InvalidUsername -> invalid username returns 400
func TestRegisterHandler_InvalidUsername(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	req := httptest.NewRequest("POST", "/register", nil)
	req.Form = url.Values{
		"email":            {"test@example.com"},
		"username":         {"!bad"},
		"password":         {"Password1!"},
		"confirm_password": {"Password1!"},
	}
	w := httptest.NewRecorder()

	RegisterHandler(store)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestRegisterHandler_InvalidPassword -> invalid password returns 400
func TestRegisterHandler_InvalidPassword(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	badPasswords := []string{
		"short1!",        // too short
		"alllowercase1!", // missing uppercase
		"ALLUPPERCASE1!", // missing lowercase
		"NoNumber!",      // missing digit
		"NoSpecial1",     // missing special char
	}

	for _, pw := range badPasswords {
		req := httptest.NewRequest("POST", "/register", nil)
		req.Form = url.Values{
			"email":            {"test@example.com"},
			"username":         {"validuser"},
			"password":         {pw},
			"confirm_password": {pw},
		}
		w := httptest.NewRecorder()

		RegisterHandler(store)(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for password %s, got %d", pw, w.Code)
		}
	}
}

// TestRegisterHandler_EmailExists -> should return 400
func TestRegisterHandler_EmailExists(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	store.CreateUser("test@example.com", "user1", "hash")

	req := httptest.NewRequest("POST", "/register", nil)
	req.Form = url.Values{
		"email":            {"test@example.com"},
		"username":         {"newuser"},
		"password":         {"Password1!"},
		"confirm_password": {"Password1!"},
	}
	w := httptest.NewRecorder()

	RegisterHandler(store)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestRegisterHandler_UsernameExists -> should return 400
func TestRegisterHandler_UsernameExists(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	store.CreateUser("test@example.com", "user1", "hash")

	req := httptest.NewRequest("POST", "/register", nil)
	req.Form = url.Values{
		"email":            {"new@example.com"},
		"username":         {"user1"},
		"password":         {"Password1!"},
		"confirm_password": {"Password1!"},
	}
	w := httptest.NewRecorder()

	RegisterHandler(store)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// UPDATED: TestRegisterHandler_Valid -> should create user and redirect to /
func TestRegisterHandler_Valid(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	req := httptest.NewRequest("POST", "/register", nil)
	req.Form = url.Values{
		"email":            {"new@example.com"},
		"username":         {"newuser"},
		"password":         {"Password1!"},
		"confirm_password": {"Password1!"},
	}
	w := httptest.NewRecorder()

	RegisterHandler(store)(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}

	// Should redirect to home now
	if w.Header().Get("Location") != "/" {
		t.Fatalf("expected redirect to '/', got %s", w.Header().Get("Location"))
	}

	// Should set session cookie
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie")
	}
}

// TestLoginHandler_Get -> should return placeholder page
func TestLoginHandler_Get(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	req := httptest.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()

	LoginHandler(store)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Login Page Placeholder") {
		t.Fatalf("expected placeholder text")
	}
}

// TestLoginHandler_MissingFields -> should return 400
func TestLoginHandler_MissingFields(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	req := httptest.NewRequest("POST", "/login", nil)
	req.Form = url.Values{
		"email":    {""},
		"password": {""},
	}
	w := httptest.NewRecorder()

	LoginHandler(store)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestLoginHandler_EmailNotFound -> should return 401
func TestLoginHandler_EmailNotFound(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	req := httptest.NewRequest("POST", "/login", nil)
	req.Form = url.Values{
		"email":    {"missing@example.com"},
		"password": {"Password1!"},
	}
	w := httptest.NewRecorder()

	LoginHandler(store)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// TestLoginHandler_WrongPassword -> should return 401
func TestLoginHandler_WrongPassword(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	hash, _ := bcrypt.GenerateFromPassword([]byte("Correct1!"), bcrypt.DefaultCost)
	store.CreateUser("test@example.com", "user1", string(hash))

	req := httptest.NewRequest("POST", "/login", nil)
	req.Form = url.Values{
		"email":    {"test@example.com"},
		"password": {"Wrong1!"},
	}
	w := httptest.NewRecorder()

	LoginHandler(store)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// TestLoginHandler_Valid -> should create session, set cookie, redirect to /
func TestLoginHandler_Valid(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	hash, _ := bcrypt.GenerateFromPassword([]byte("Correct1!"), bcrypt.DefaultCost)
	uid, _ := store.CreateUser("test@example.com", "user1", string(hash))

	req := httptest.NewRequest("POST", "/login", nil)
	req.Form = url.Values{
		"email":    {"test@example.com"},
		"password": {"Correct1!"},
	}
	w := httptest.NewRecorder()

	LoginHandler(store)(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie")
	}

	if cookies[0].Name != "session" {
		t.Fatalf("expected session cookie, got %s", cookies[0].Name)
	}

	// Ensure session exists in DB
	_, err := store.GetSession(cookies[0].Value)
	if err != nil {
		t.Fatalf("expected session in DB, got %v", err)
	}

	// Ensure old sessions were deleted
	sessions, _ := store.DB.Query("SELECT token FROM sessions WHERE user_id = ?", uid)
	defer sessions.Close()
	count := 0
	for sessions.Next() {
		count++
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 active session, got %d", count)
	}
}

// TestLogoutHandler_Get -> should return 405
func TestLogoutHandler_Get(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	req := httptest.NewRequest("GET", "/logout", nil)
	w := httptest.NewRecorder()

	LogoutHandler(store)(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// UPDATED: TestLogoutHandler_NoCookie -> should redirect to /
func TestLogoutHandler_NoCookie(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	req := httptest.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()

	LogoutHandler(store)(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}

	if w.Header().Get("Location") != "/" {
		t.Fatalf("expected redirect to '/', got %s", w.Header().Get("Location"))
	}
}

// UPDATED: TestLogoutHandler_Valid -> should delete session, clear cookie, redirect to /
func TestLogoutHandler_Valid(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, _ := db.InitializeDatabase(dbPath)
	defer conn.Close()

	store := db.NewStore(conn)

	uid, _ := store.CreateUser("test@example.com", "user1", "hash")
	token := "tok123"
	expires := time.Now().Add(1 * time.Hour)
	store.CreateSession(uid, token, expires)

	req := httptest.NewRequest("POST", "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	LogoutHandler(store)(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}

	if w.Header().Get("Location") != "/" {
		t.Fatalf("expected redirect to '/', got %s", w.Header().Get("Location"))
	}

	// Session should be deleted
	_, err := store.GetSession(token)
	if err != db.ErrNotFound {
		t.Fatalf("expected session to be deleted, got %v", err)
	}

	// Cookie should be cleared
	cookies := w.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Value != "" {
		t.Fatalf("expected cleared cookie")
	}
}
