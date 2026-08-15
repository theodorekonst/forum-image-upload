package handlers

import (
	"encoding/json"
	"errors"
	"forum/config"
	"forum/db"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GitHubLoginHandler starts the OAuth flow by redirecting the user to GitHub.
func GitHubLoginHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		params := url.Values{}
		params.Add("client_id", config.GitHubClientID)
		params.Add("redirect_uri", "http://localhost:8080/auth/github/callback")
		params.Add("scope", "user:email")

		authURL := "https://github.com/login/oauth/authorize?" + params.Encode()

		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

// GitHubCallbackHandler handles the OAuth callback from GitHub.
func GitHubCallbackHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Read ?code from query
		code := r.URL.Query().Get("code")
		if code == "" {
			RenderError(w, http.StatusBadRequest, "Missing authorization code.")
			return
		}

		// 2. Exchange code for access token
		clientID := config.GitHubClientID
		clientSecret := config.GitHubClientSecret

		tokenReqBody := url.Values{}
		tokenReqBody.Set("client_id", clientID)
		tokenReqBody.Set("client_secret", clientSecret)
		tokenReqBody.Set("code", code)
		tokenReqBody.Set("redirect_uri", "http://localhost:8080/auth/github/callback")

		// Fix during testing: GitHub requires Accept: application/json
		req, _ := http.NewRequest(
			"POST",
			"https://github.com/login/oauth/access_token",
			strings.NewReader(tokenReqBody.Encode()),
		)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		tokenResp, err := http.DefaultClient.Do(req)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to exchange code for token.")
			return
		}
		defer tokenResp.Body.Close()

		var tokenData struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			Scope       string `json:"scope"`
		}

		if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to parse token response.")
			return
		}

		if tokenData.AccessToken == "" {
			RenderError(w, http.StatusUnauthorized, "GitHub did not return an access token.")
			return
		}

		// 3. Fetch GitHub user info
		reqUser, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
		reqUser.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)

		userResp, err := http.DefaultClient.Do(reqUser)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to fetch GitHub user.")
			return
		}
		defer userResp.Body.Close()

		var ghUser struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
			Email string `json:"email"`
		}

		if err := json.NewDecoder(userResp.Body).Decode(&ghUser); err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to parse GitHub user.")
			return
		}

		// 4. Check if user exists or create new OAuth user
		provider := "github"
		oauthID := strconv.FormatInt(ghUser.ID, 10)

		user, err := store.GetUserByOAuth(provider, oauthID)
		if errors.Is(err, db.ErrNotFound) {
			userID, err := store.CreateOAuthUser(ghUser.Email, ghUser.Login, provider, oauthID)
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
