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

var oauthHTTPClient = &http.Client{Timeout: 10 * time.Second}

func generateOAuthState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func setOAuthStateCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})
}

func clearOAuthStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func validateOAuthState(r *http.Request, expected string) bool {
	cookie, err := r.Cookie(oauthStateCookieName)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(expected)) == 1
}

func setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
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

func IssueSession(w http.ResponseWriter, store *db.Store, userID int64) (string, error) {
	if err := store.DeleteSessionsByUserID(userID); err != nil {
		return "", err
	}

	token := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	if _, err := store.CreateSession(userID, token, expiresAt); err != nil {
		return "", err
	}

	setSessionCookie(w, token, expiresAt)
	return token, nil
}

func CreateOrLinkOAuthUser(store *db.Store, provider, oauthID, email, username string) (*db.User, error) {
	user, err := store.GetUserByOAuth(provider, oauthID)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return nil, err
	}

	emailToUse := email
	if emailToUse == "" {
		emailToUse = oauthID + "@" + provider + ".invalid"
	}

	userID, err := store.CreateOAuthUser(emailToUse, username, provider, oauthID)
	if err == nil {
		return store.GetUserByID(userID)
	}

	existing, e2 := store.GetUserByEmail(emailToUse)
	if e2 == nil {
		if err2 := store.SetUserOAuth(existing.ID, provider, oauthID); err2 != nil {
			return nil, err2
		}
		return existing, nil
	}

	return nil, err
}

func CreateSessionAndSetCookie(w http.ResponseWriter, store *db.Store, user *db.User) error {
	_, err := IssueSession(w, store, user.ID)
	return err
}
