package handlers

import (
	"errors"
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
			confirmPassword := r.FormValue("confirm_password")

			registerError := func(field, msg string) {
				w.WriteHeader(http.StatusBadRequest)
				data := map[string]any{
					"Error":      msg,
					"Email":      email,
					"Username":   username,
					"FirstError": field,
				}
				data[field+"Error"] = msg
				renderPage(w, "register.html", data)
			}

			// 2. Validate input
			if email == "" {
				registerError("Email", "Email is required.")
				return
			}
			if username == "" {
				registerError("Username", "Username is required.")
				return
			}
			if password == "" {
				registerError("Password", "Password is required.")
				return
			}
			if confirmPassword == "" {
				registerError("ConfirmPassword", "Please confirm your password.")
				return
			}

			if password != confirmPassword {
				registerError("ConfirmPassword", "Passwords do not match.")
				return
			}

			if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
				registerError("Email", "Enter a valid email address.")
				return
			}
			usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
			if !usernameRegex.MatchString(username) {
				registerError("Username", "Use 3–20 letters, numbers, or underscores.")
				return
			}
			if len(password) < 8 || len(password) > 72 {
				registerError("Password", "Use between 8 and 72 characters.")
				return
			}
			if !regexp.MustCompile(`[A-Z]`).MatchString(password) ||
				!regexp.MustCompile(`[a-z]`).MatchString(password) ||
				!regexp.MustCompile(`[0-9]`).MatchString(password) ||
				!regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
				registerError("Password", "Include uppercase, lowercase, number, and special character.")
				return
			}

			// 3. Check if email exists
			_, err := store.GetUserByEmail(email)
			if err == nil {
				registerError("Email", "This email is already in use.")
				return
			}
			if !errors.Is(err, db.ErrNotFound) {
				log.Println("error checking email:", err)
				RenderError(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
				return
			}

			// 4. Check if username exists
			_, err = store.GetUserByUsername(username)
			if err == nil {
				registerError("Username", "This username is already taken.")
				return
			}
			if !errors.Is(err, db.ErrNotFound) {
				log.Println("error checking username:", err)
				RenderError(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
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
			newUserID, err := store.CreateUser(email, username, string(hash))
			if err != nil {
				log.Println("error creating user:", err)
				RenderError(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
				return
			}

			// 7. Auto-login after successful registration
			token := uuid.New().String()
			expiresAt := time.Now().Add(24 * time.Hour)

			_, err = store.CreateSession(newUserID, token, expiresAt)
			if err != nil {
				log.Println("error creating session:", err)
				RenderError(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    token,
				Path:     "/",
				Expires:  expiresAt,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})

			// Redirect to home
			http.Redirect(w, r, "/", http.StatusSeeOther)
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

			loginError := func(status int, field, msg string) {
				w.WriteHeader(status)
				data := map[string]any{
					"Email":      email,
					"FirstError": field,
				}
				if field == "Credentials" {
					data["CredentialsError"] = msg
					data["FirstError"] = "Password"
				} else {
					data[field+"Error"] = msg
				}
				renderPage(w, "login.html", data)
			}

			// 2. Basic empty check
			if email == "" {
				loginError(http.StatusBadRequest, "Email", "Email is required.")
				return
			}
			if password == "" {
				loginError(http.StatusBadRequest, "Password", "Password is required.")
				return
			}

			// 3. Verify email exists
			user, err := store.GetUserByEmail(email)
			if errors.Is(err, db.ErrNotFound) {
				loginError(http.StatusUnauthorized, "Credentials", "Email or password is incorrect.")
				return
			}
			if err != nil {
				log.Println("error loading user:", err)
				RenderError(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
				return
			}

			// 4. Verify password matches stored hash
			err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
			if err != nil {
				loginError(http.StatusUnauthorized, "Credentials", "Email or password is incorrect.")
				return
			}

			// 5. Session & Cookie Handling
			// Delete old sessions
			if err := store.DeleteSessionsByUserID(user.ID); err != nil {
				log.Println("error deleting old sessions:", err)
				RenderError(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
				return
			}

			// Generate new token
			token := uuid.New().String()

			// Create session in database
			expiresAt := time.Now().Add(24 * time.Hour)
			_, err = store.CreateSession(user.ID, token, expiresAt)
			if err != nil {
				log.Println("error creating session:", err)
				RenderError(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
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
			RenderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
			return
		}

		// POST -> process logout
		if r.Method == http.MethodPost {

			// 1. Read session cookie
			cookie, err := r.Cookie("session")
			if err != nil {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			// 2. Delete session from database
			err = store.DeleteSession(cookie.Value)
			if err != nil {
				log.Println("error deleting session", err)
			}

			// 3. Clear cookie
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

			// 4. Redirect to home
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Unsupported method
		RenderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
	}
}
