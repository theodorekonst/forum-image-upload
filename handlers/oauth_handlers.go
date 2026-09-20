package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"forum/db"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const oauthStateCookieName = "oauth_state"

// Shared HTTP client for OAuth requests (10s timeout)
var oauthHTTPClient = &http.Client{Timeout: 10 * time.Second}

// generateOAuthState creates a random 16‑byte hex string used as the OAuth state.
func generateOAuthState() (string, error) {

	// 1. Generate random bytes
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// 2. Convert to hex string
	return hex.EncodeToString(bytes), nil
}

// setOAuthStateCookie stores the OAuth state in a short‑lived cookie.
func setOAuthStateCookie(w http.ResponseWriter, state string) {

	// 1. Write cookie
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})
}

// clearOAuthStateCookie removes the OAuth state cookie.
func clearOAuthStateCookie(w http.ResponseWriter) {

	// 1. Overwrite cookie with expired value
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// validateOAuthState compares the stored cookie state with the expected value.
func validateOAuthState(r *http.Request, expected string) bool {

	// 1. Read cookie
	cookie, err := r.Cookie(oauthStateCookieName)
	if err != nil {
		return false
	}

	// 2. Constant‑time comparison
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(expected)) == 1
}

// setSessionCookie writes the session cookie with expiration.
func setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {

	// 1. Write cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearSessionCookie removes the session cookie.
func clearSessionCookie(w http.ResponseWriter) {

	// 1. Overwrite cookie with expired value
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// IssueSession deletes old sessions, creates a new one, and sets the cookie.
func IssueSession(w http.ResponseWriter, store *db.Store, userID int64) (string, error) {

	// 1. Remove existing sessions for this user
	if err := store.DeleteSessionsByUserID(userID); err != nil {
		return "", err
	}

	// 2. Generate new session token
	token := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	// 3. Store session in database
	if _, err := store.CreateSession(userID, token, expiresAt); err != nil {
		return "", err
	}

	// 4. Set session cookie
	setSessionCookie(w, token, expiresAt)
	return token, nil
}

// CreateOrLinkOAuthUser finds or creates a user associated with an OAuth provider.
func CreateOrLinkOAuthUser(store *db.Store, provider, oauthID, email, username string) (*db.User, error) {

	// 1. Check if OAuth user already exists
	user, err := store.GetUserByOAuth(provider, oauthID)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return nil, err
	}

	// 2. Determine email to use (fallback if missing)
	emailToUse := email
	if emailToUse == "" {
		emailToUse = oauthID + "@" + provider + ".invalid"
	}

	// 3. Create new OAuth user
	userID, err := store.CreateOAuthUser(emailToUse, username, provider, oauthID)
	if err == nil {
		return store.GetUserByID(userID)
	}

	// 4. If email exists, link OAuth provider to existing user
	existing, e2 := store.GetUserByEmail(emailToUse)
	if e2 == nil {
		if err2 := store.SetUserOAuth(existing.ID, provider, oauthID); err2 != nil {
			return nil, err2
		}
		return existing, nil
	}

	// 5. Failure
	return nil, err
}

// CreateSessionAndSetCookie issues a session for the given user.
func CreateSessionAndSetCookie(w http.ResponseWriter, store *db.Store, user *db.User) error {

	// 1. Issue session
	_, err := IssueSession(w, store, user.ID)
	return err
}
