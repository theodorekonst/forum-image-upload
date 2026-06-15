package handlers

import (
    "net/http"
    "forum/db"
)

// RegisterHandler handles user registration.
func RegisterHandler(store *db.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // TODO
    }
}

// LoginHandler handles user login.
func LoginHandler(store *db.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // TODO
    }
}

// LogoutHandler handles user logout.
func LogoutHandler(store *db.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // TODO
    }
}
