package handlers

import (
	"errors"
	"forum/db"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// CreateOrLinkOAuthUser ensures a user exists for the given OAuth provider/id.
func CreateOrLinkOAuthUser(store *db.Store, provider, oauthID, email, username string) (*db.User, error) {

	// Check if OAuth user already exists
	user, err := store.GetUserByOAuth(provider, oauthID)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return nil, err
	}

	// Prepare fallback email if provider does not supply one
	emailToUse := email
	if emailToUse == "" {
		emailToUse = oauthID + "@" + provider + ".invalid"
	}

	// Attempt to create new OAuth user
	userID, err := store.CreateOAuthUser(emailToUse, username, provider, oauthID)
	if err == nil {
		return store.GetUserByID(userID)
	}

	// If creation failed, try linking to an existing user with the same email
	existing, e2 := store.GetUserByEmail(emailToUse)
	if e2 == nil {
		if err2 := store.SetUserOAuth(existing.ID, provider, oauthID); err2 != nil {
			return nil, err2
		}
		return existing, nil
	}

	return nil, err
}

// CreateSessionAndSetCookie creates a session and sets the session cookie.
func CreateSessionAndSetCookie(w http.ResponseWriter, store *db.Store, user *db.User) error {

	// Generate session token
	token := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	// Store session in database
	_, err := store.CreateSession(user.ID, token, expiresAt)
	if err != nil {
		return err
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
	})

	return nil
}
