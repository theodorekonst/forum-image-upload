package middleware

import (
    "net/http"
    "forum/db"
)

// GetUserFromRequest extracts the logged-in user from the session cookie.
func GetUserFromRequest(store *db.Store, r *http.Request) (*db.User, error) {
    // TODO
    return nil, nil
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
