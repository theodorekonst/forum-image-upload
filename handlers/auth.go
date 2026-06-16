package handlers

import (
	"forum/db"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// RegisterHandler handles user registration.
func RegisterHandler(store *db.Store) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {

		// GET -> show placeholder page
		if req.Method == http.MethodGet {
			wr.Write([]byte("Register Page Placeholder"))
			return
		}

		// POST -> process registration
		if req.Method == http.MethodPost {

			// 1. Read form values
			email := req.FormValue("email")
			username := req.FormValue("username")
			password := req.FormValue("password")

			// 2. Validate input
			// Basic empty check
			if email == "" || username == "" || password == "" {
				http.Error(wr, "please fill all fields", http.StatusBadRequest)
				return
			}
			// Email format check
			if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
				http.Error(wr, "invalid email format", http.StatusBadRequest)
				return
			}
			// Username: 3–20 chars, letters/numbers/underscore
			usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
			if !usernameRegex.MatchString(username) {
				http.Error(wr, "username must be 3-20 characters and may include letters, numbers and underscore", http.StatusBadRequest)
				return
			}
			// Password: 8–72 chars, uppercase, lowercase, digit, special character
			if len(password) < 8 || len(password) > 72 {
				http.Error(wr, "password must be between 8 and 72 characters", http.StatusBadRequest)
				return
			}

			if !regexp.MustCompile(`[A-Z]`).MatchString(password) ||
				!regexp.MustCompile(`[a-z]`).MatchString(password) ||
				!regexp.MustCompile(`[0-9]`).MatchString(password) ||
				!regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
				http.Error(wr, "password must include uppercase, lowercase, number and special character", http.StatusBadRequest)
				return
			}

			// 3. Check if email exists
			_, err := store.GetUserByEmail(email)
			if err == nil {
				http.Error(wr, "email already in use", http.StatusBadRequest)
				return
			}

			// 4. Hash password
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				log.Println("error hashing password:", err)
				http.Error(wr, "internal server error", http.StatusInternalServerError)
				return
			}

			// 5. Create user
			_, err = store.CreateUser(email, username, string(hash))
			if err != nil {
				log.Println("error creating user:", err)
				http.Error(wr, "internal server error", http.StatusInternalServerError)
				return
			}

			// 6. Redirect to login
			http.Redirect(wr, req, "/login", http.StatusSeeOther)
			return
		}

		// Unsupported method
		http.Error(wr, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// LoginHandler handles user login.
func LoginHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// GET -> show placeholder page
		if r.Method == http.MethodGet {
			w.Write([]byte("Login Page Placeholder"))
			return
		}

		// POST -> process login
		if r.Method == http.MethodPost {

			// 1. Read form values
			email := r.FormValue("email")
			password := r.FormValue("password")

			// 2. Basic empty check
			if email == "" || password == "" {
				http.Error(w, "please fill all fields", http.StatusBadRequest)
				return
			}

			// 3. Verify email exists (err != nil means user not found by the way)
			user, err := store.GetUserByEmail(email)
			if err != nil {
				http.Error(w, "invalid credentials", http.StatusBadRequest)
				return
			}

			// 4. Verify password matches stored hash
			err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
			if err != nil {
				http.Error(w, "invalid credentials", http.StatusBadRequest)
				return
			}

			// 5. Successful login (sessions will be added later)
			w.Write([]byte("Login successful (session not implemented yet)"))
			return
		}

		// Unsupported method
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// LogoutHandler handles user logout.
func LogoutHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO
	}
}
