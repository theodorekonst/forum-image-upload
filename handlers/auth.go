package handlers

import (
	"forum/db"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Temporary placeholder for home page (testing purposes)
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Home Page Placeholder"))
}

// RegisterHandler handles user registration.
func RegisterHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// GET -> show placeholder page
		if r.Method == http.MethodGet {
			w.Write([]byte("Register Page Placeholder"))
			return
		}

		// POST -> process registration
		if r.Method == http.MethodPost {

			// 1. Read form values
			email := r.FormValue("email")
			username := r.FormValue("username")
			password := r.FormValue("password")

			// 2. Validate input
			// Basic empty check
			if email == "" || username == "" || password == "" {
				http.Error(w, "please fill all fields", http.StatusBadRequest)
				return
			}
			// Email format check
			if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
				http.Error(w, "invalid email format", http.StatusBadRequest)
				return
			}
			// Username: 3–20 chars, letters/numbers/underscore
			usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
			if !usernameRegex.MatchString(username) {
				http.Error(w, "username must be 3-20 characters and may include letters, numbers and underscore", http.StatusBadRequest)
				return
			}
			// Password: 8–72 chars, uppercase, lowercase, digit, special character
			if len(password) < 8 || len(password) > 72 {
				http.Error(w, "password must be between 8 and 72 characters", http.StatusBadRequest)
				return
			}

			if !regexp.MustCompile(`[A-Z]`).MatchString(password) ||
				!regexp.MustCompile(`[a-z]`).MatchString(password) ||
				!regexp.MustCompile(`[0-9]`).MatchString(password) ||
				!regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
				http.Error(w, "password must include uppercase, lowercase, number and special character", http.StatusBadRequest)
				return
			}

			// 3. Check if email exists
			_, err := store.GetUserByEmail(email)
			if err == nil {
				http.Error(w, "email already in use", http.StatusBadRequest)
				return
			}

			// 4. Check if username exists
			_, err = store.GetUserByUsername(username)
			if err == nil {
				http.Error(w, "username already exists", http.StatusBadRequest)
				return
			}

			// 5. Hash password
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				log.Println("error hashing password:", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			// 6. Create user
			_, err = store.CreateUser(email, username, string(hash))
			if err != nil {
				log.Println("error creating user:", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			// 7. Redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Unsupported method
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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

			// 5. Session & Cookie Handling
			// Delete old sessions
			store.DeleteSessionsByUserID(user.ID)

			// Generate new token
			token := uuid.New().String()

			// Create session in database
			expiresAt := time.Now().Add(24 * time.Hour)
			_, err = store.CreateSession(user.ID, token, expiresAt)
			if err != nil {
				log.Println("error creating session:", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			// Set cookie
			http.SetCookie(w, &http.Cookie{
				Name:  "session",
				Value: token,
				Path:  "/",
				// Domain -> in production when there are subdomains
				Expires: expiresAt,
				// MaxAge -> alternative to Expires (MaxAge: 86400)
				Secure:   false, // in production, true
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})

			// 6. Redirect to home
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}

		// Unsupported method
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// LogoutHandler handles user logout.
func LogoutHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// GET -> method not allowed (CSRF protection)
		if r.Method == http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// POST -> process logout
		if r.Method == http.MethodPost {

			// 1. Read session cookie
			cookie, err := r.Cookie("session")
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// 2. Delete session from database
			err = store.DeleteSession(cookie.Value)
			if err != nil {
				log.Println("error deleting session", err)
			}

			// 3. Clear cookie in browser
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0), // expires immediately
				MaxAge:   -1,              // deletes immediately
				Secure:   false,           // in production, true
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})

			// 4. Redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}

		// Unsupported method
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
