package handlers

import (
	"encoding/json"
	"errors"
	"forum/config"
	"forum/db"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DiscordLoginHandler starts the OAuth flow by redirecting the user to Discord.
func DiscordLoginHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		params := url.Values{}
		params.Add("client_id", config.DiscordClientID)
		params.Add("redirect_uri", config.DiscordRedirectURL)
		params.Add("response_type", "code")
		params.Add("scope", "identify email")

		authURL := config.DiscordAuthURL + "?" + params.Encode()

		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

// DiscordCallbackHandler handles the OAuth callback from Discord.
func DiscordCallbackHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Read ?code from query
		code := r.URL.Query().Get("code")
		if code == "" {
			RenderError(w, http.StatusBadRequest, "Missing authorization code.")
			return
		}

		// 2. Exchange code for access token
		tokenReqBody := url.Values{}
		tokenReqBody.Set("client_id", config.DiscordClientID)
		tokenReqBody.Set("client_secret", config.DiscordClientSecret)
		tokenReqBody.Set("code", code)
		tokenReqBody.Set("grant_type", "authorization_code")
		tokenReqBody.Set("redirect_uri", config.DiscordRedirectURL)

		req, _ := http.NewRequest(
			"POST",
			config.DiscordTokenURL,
			strings.NewReader(tokenReqBody.Encode()),
		)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		tokenResp, err := http.DefaultClient.Do(req)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to exchange code for token.")
			return
		}
		defer tokenResp.Body.Close()

		var tokenData struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int    `json:"expires_in"`
			Scope       string `json:"scope"`
		}

		if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to parse token response.")
			return
		}

		if tokenData.AccessToken == "" {
			RenderError(w, http.StatusUnauthorized, "Discord did not return an access token.")
			return
		}

		// 3. Fetch Discord user info
		reqUser, _ := http.NewRequest("GET", config.DiscordUserInfoURL, nil)
		reqUser.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)

		userResp, err := http.DefaultClient.Do(reqUser)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to fetch Discord user.")
			return
		}
		defer userResp.Body.Close()

		var dUser struct {
			ID            string `json:"id"`
			Username      string `json:"username"`
			Discriminator string `json:"discriminator"`
			Email         string `json:"email"`
		}

		if err := json.NewDecoder(userResp.Body).Decode(&dUser); err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to parse Discord user.")
			return
		}

		// 4. Check if user exists or create new OAuth user
		provider := "discord"
		oauthID := dUser.ID

		user, err := store.GetUserByOAuth(provider, oauthID)
		if errors.Is(err, db.ErrNotFound) {
			// Discord usernames are "name#1234"
			fullName := dUser.Username
			if dUser.Discriminator != "" {
				fullName = dUser.Username + "#" + dUser.Discriminator
			}

			userID, err := store.CreateOAuthUser(dUser.Email, fullName, provider, oauthID)
			if err != nil {
				RenderError(w, http.StatusInternalServerError, "Failed to create OAuth user.")
				return
			}
			user, _ = store.GetUserByID(userID)
		}

		// 5. Create session and set cookie
		token := uuid.New().String()
		expiresAt := time.Now().Add(24 * time.Hour)

		_, err = store.CreateSession(user.ID, token, expiresAt)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to create session.")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			Expires:  expiresAt,
			HttpOnly: true,
		})

		// 6. Redirect to home
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
