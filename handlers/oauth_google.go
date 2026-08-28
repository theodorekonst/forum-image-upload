package handlers

import (
	"encoding/json"
	"forum/config"
	"forum/db"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// GoogleLoginHandler begins the OAuth flow by redirecting the user to Google's authorization page.
func GoogleLoginHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Generate OAuth state
		state, err := generateOAuthState()
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to start Google login.")
			return
		}
		setOAuthStateCookie(w, state)

		// 2. Build Google authorization URL
		params := url.Values{}
		params.Add("client_id", config.GoogleClientID)
		params.Add("redirect_uri", config.RedirectURL("/auth/google/callback"))
		params.Add("response_type", "code")
		params.Add("scope", "email profile")
		params.Add("access_type", "offline")
		params.Add("prompt", "consent")
		params.Add("state", state)

		authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()

		// 3. Redirect user to Google OAuth page
		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

// GoogleCallbackHandler processes the OAuth callback after Google redirects back with a code.
func GoogleCallbackHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Read authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			RenderError(w, http.StatusBadRequest, "Missing authorization code.")
			return
		}

		// 2. Validate OAuth state
		state := r.URL.Query().Get("state")
		if !validateOAuthState(r, state) {
			RenderError(w, http.StatusBadRequest, "Invalid OAuth state.")
			return
		}
		clearOAuthStateCookie(w)

		// 3. Exchange code for access token
		tokenReqBody := url.Values{}
		tokenReqBody.Set("client_id", config.GoogleClientID)
		tokenReqBody.Set("client_secret", config.GoogleClientSecret)
		tokenReqBody.Set("code", code)
		tokenReqBody.Set("grant_type", "authorization_code")
		tokenReqBody.Set("redirect_uri", config.RedirectURL("/auth/google/callback"))

		req, _ := http.NewRequest(
			"POST",
			"https://oauth2.googleapis.com/token",
			strings.NewReader(tokenReqBody.Encode()),
		)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		tokenResp, err := oauthHTTPClient.Do(req)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to exchange code for token.")
			return
		}
		defer tokenResp.Body.Close()

		// 4. Parse token response
		var tokenData struct {
			AccessToken string `json:"access_token"`
			IDToken     string `json:"id_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int    `json:"expires_in"`
		}

		if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to parse token response.")
			return
		}

		if tokenData.AccessToken == "" {
			RenderError(w, http.StatusUnauthorized, "Google did not return an access token.")
			return
		}

		// 5. Retrieve Google user profile
		reqUser, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
		reqUser.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)

		userResp, err := oauthHTTPClient.Do(reqUser)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to fetch Google user.")
			return
		}
		defer userResp.Body.Close()

		// 6. Parse user profile
		var gUser struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}

		if err := json.NewDecoder(userResp.Body).Decode(&gUser); err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to parse Google user.")
			return
		}

		// 7. Create or link OAuth user
		provider := "google"
		oauthID := gUser.ID

		user, err := CreateOrLinkOAuthUser(store, provider, oauthID, gUser.Email, gUser.Name)
		if err != nil {
			log.Println("error creating or linking oauth user:", err)
			RenderError(w, http.StatusInternalServerError, "Failed to create or link OAuth user.")
			return
		}

		// 8. Create session and set cookie
		if err := CreateSessionAndSetCookie(w, store, user); err != nil {
			log.Println("error creating session:", err)
			RenderError(w, http.StatusInternalServerError, "Failed to create session.")
			return
		}

		// 9. Redirect user to home page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
