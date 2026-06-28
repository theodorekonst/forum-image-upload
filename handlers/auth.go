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

// RegisterHandler handles user registration.
func RegisterHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// GET -> show registration form
		if r.Method == http.MethodGet {
			if err := renderPage(w, "register.html", nil); err != nil {
				RenderError(w, http.StatusInternalServerError, "Could not render the registration page.")
			}
			return
		}

		// POST -> process registration
		if r.Method == http.MethodPost {

			// 1. Read form values
			email := r.FormValue("email")
			username := r.FormValue("username")
			password := r.FormValue("password")

			registerError := func(msg string) {
				w.WriteHeader(http.StatusBadRequest)
				renderPage(w, "register.html", map[string]any{
					"Error":    msg,
					"Email":    email,
					"Username": username,
				})
			}

			// 2. Validate input
			if email == "" || username == "" || password == "" {
				registerError("Please fill all fields.")
				return
			}
			if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
				registerError("Invalid email format.")
				return
			}
			usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
			if !usernameRegex.MatchString(username) {
				registerError("Username must be 3–20 characters: letters, numbers, underscore only.")
				return
			}
			if len(password) < 8 || len(password) > 72 {
				registerError("Password must be between 8 and 72 characters.")
				return
			}
			if !regexp.MustCompile(`[A-Z]`).MatchString(password) ||
				!regexp.MustCompile(`[a-z]`).MatchString(password) ||
				!regexp.MustCompile(`[0-9]`).MatchString(password) ||
				!regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
				registerError("Password must include uppercase, lowercase, number and special character.")
				return
			}

			// 3. Check if email exists
			_, err := store.GetUserByEmail(email)
			if err == nil {
				registerError("This email is already in use.")
				return
			}

			// 4. Check if username exists
			_, err = store.GetUserByUsername(username)
			if err == nil {
				registerError("This username is already taken.")
				return
			}

			// 5. Hash password
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				log.Println("error hashing password:", err)
				RenderError(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
				return
			}

			// 6. Create user
			_, err = store.CreateUser(email, username, string(hash))
			if err != nil {
				log.Println("error creating user:", err)
				RenderError(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
				return
			}

			// 7. Redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		RenderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
	}
}

// LoginHandler handles user login.
func LoginHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// GET -> show login form
		if r.Method == http.MethodGet {
			if err := renderPage(w, "login.html", nil); err != nil {
				RenderError(w, http.StatusInternalServerError, "Could not render the login page.")
			}
			return
		}

		// POST -> process login
		if r.Method == http.MethodPost {

			// 1. Read form values
			email := r.FormValue("email")
			password := r.FormValue("password")

			loginError := func(msg string) {
				w.WriteHeader(http.StatusUnauthorized)
				renderPage(w, "login.html", map[string]any{
					"Error": msg,
					"Email": email,
				})
			}

			// 2. Basic empty check
			if email == "" || password == "" {
				loginError("Please fill all fields.")
				return
			}

			// 3. Verify email exists
			user, err := store.GetUserByEmail(email)
			if err != nil {
				loginError("Invalid email or password.")
				return
			}

			// 4. Verify password matches stored hash
			err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
			if err != nil {
				loginError("Invalid email or password.")
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
			return
		}

		RenderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
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
			return
		}

		// Unsupported method
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
