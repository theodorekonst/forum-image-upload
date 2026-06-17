package middleware

import (
	"forum/db"
	"net/http"
	"time"
)

// GetUserFromRequest extracts the logged-in user from the session cookie.
func GetUserFromRequest(store *db.Store, r *http.Request) (*db.User, error) {

	// 1. Read session cookie
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil, err
	}

	// 2. Look up session in database
	session, err := store.GetSession(cookie.Value)
	if err != nil {
		return nil, err
	}

	// 3. Check expiration
	if time.Now().After(session.ExpiresAt) {
		store.DeleteSession(session.Token)
		return nil, db.ErrNotFound
	}

	// 4. Load user
	user, err := store.GetUserByID(session.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// RequireAuth ensures only logged-in users can access a route.
func RequireAuth(store *db.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})
}

// RequireGuest ensures logged-in users cannot access login/register.
func RequireGuest(store *db.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})
}
